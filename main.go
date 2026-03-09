package main

import (
	"fmt"
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// AppState manages the UI widgets
type AppState struct {
	MainWindow    *walk.MainWindow
	MethodCB      *walk.ComboBox
	UrlLE         *walk.LineEdit
	SendBtn       *walk.PushButton
	HeadersTE     *walk.TextEdit
	BodyTE        *walk.TextEdit
	StatusLbl     *walk.Label
	RespBodyTE    *walk.TextEdit
	EndpointsList *walk.ListBox

	Endpoints       []Endpoint
	EndpointStrings []string
	HttpMethods     []string
}

func main() {
	state := &AppState{
		HttpMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
	}

	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}

	if err := (MainWindow{
		AssignTo:   &state.MainWindow,
		Title:      "Postlang",
		MinSize:    Size{Width: 900, Height: 600},
		Font:       Font{Family: "Segoe UI", PointSize: 10},
		Background: bgBrush,
		Layout:     HBox{MarginsZero: true},
		MenuItems:  state.buildMenuBar(),
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					state.buildLeftNav(),
					state.buildRightPane(),
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	state.applyDarkTheme()
	state.MainWindow.Run()
}

func (s *AppState) applyDarkTheme() {
	bg, _ := walk.NewSolidColorBrush(walk.RGB(40, 44, 52))
	inBg, _ := walk.NewSolidColorBrush(walk.RGB(33, 37, 43))
	fg := walk.RGB(220, 220, 220)

	s.MainWindow.SetBackground(bg)
	
	if s.EndpointsList != nil {
		s.EndpointsList.SetBackground(inBg)
	}
	if s.UrlLE != nil {
		s.UrlLE.SetBackground(inBg)
		s.UrlLE.SetTextColor(fg)
	}
	if s.HeadersTE != nil {
		s.HeadersTE.SetBackground(inBg)
		s.HeadersTE.SetTextColor(fg)
	}
	if s.BodyTE != nil {
		s.BodyTE.SetBackground(inBg)
		s.BodyTE.SetTextColor(fg)
	}
	if s.RespBodyTE != nil {
		s.RespBodyTE.SetBackground(inBg)
		s.RespBodyTE.SetTextColor(fg)
	}
	if s.StatusLbl != nil {
		s.StatusLbl.SetBackground(bg)
		s.StatusLbl.SetTextColor(fg)
	}
}

func (s *AppState) buildMenuBar() []MenuItem {
	return []MenuItem{
		Menu{
			Text: "&File",
			Items: []MenuItem{
				Action{
					Text: "Import OpenAPI Spec...",
					OnTriggered: func() {
						dlg := new(walk.FileDialog)
						dlg.Title = "Import OpenAPI Spec"
						dlg.Filter = "OpenAPI Files (*.json;*.yaml;*.yml)|*.json;*.yaml;*.yml|All Files (*.*)|*.*"

						if ok, err := dlg.ShowOpen(s.MainWindow); err != nil {
							return
						} else if !ok {
							return
						}

						eps, err := parseOpenAPI(dlg.FilePath)
						if err != nil {
							walk.MsgBox(s.MainWindow, "Error Loading OpenAPI", err.Error(), walk.MsgBoxIconError)
							return
						}

						s.Endpoints = eps
						s.EndpointStrings = make([]string, len(s.Endpoints))
						for i, e := range s.Endpoints {
							s.EndpointStrings[i] = e.DisplayName()
						}

						s.EndpointsList.SetModel(s.EndpointStrings)
					},
				},
			},
		},
	}
}

func (s *AppState) buildLeftNav() Widget {
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	return Composite{
		Layout:     VBox{MarginsZero: true},
		MinSize:    Size{Width: 250, Height: 0},
		Background: bgBrush,
		Children: []Widget{
			Label{
				Text:       "API Endpoints (Import from File)",
				TextColor:  walk.RGB(220, 220, 220),
				Background: bgBrush,
			},
			ListBox{
				AssignTo: &s.EndpointsList,
				Model:    s.EndpointStrings,
				OnCurrentIndexChanged: func() {
					idx := s.EndpointsList.CurrentIndex()
					if idx >= 0 && idx < len(s.Endpoints) {
						ep := s.Endpoints[idx]
						
						for i, m := range s.HttpMethods {
							if m == ep.Method {
								s.MethodCB.SetCurrentIndex(i)
								break
							}
						}

						fullURL := ep.Path
						if ep.BaseURL != "" {
							fullURL = ep.BaseURL + ep.Path
						} else {
							fullURL = "http://localhost:8080" + ep.Path
						}
						s.UrlLE.SetText(fullURL)

						if ep.Headers != "" {
							s.HeadersTE.SetText(ep.Headers)
						} else {
							s.HeadersTE.SetText("")
						}

						if ep.Body != "" {
							s.BodyTE.SetText(ep.Body)
						} else {
							s.BodyTE.SetText("")
						}
					}
				},
			},
		},
	}
}

func (s *AppState) buildRightPane() Widget {
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	return Composite{
		Layout:     VBox{},
		Background: bgBrush,
		Children: []Widget{
			s.buildTopBar(),
			VSplitter{
				Children: []Widget{
					s.buildRequestTabs(),
					s.buildResponseSection(),
				},
			},
		},
	}
}

func (s *AppState) buildTopBar() Widget {
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	return Composite{
		Layout:     HBox{MarginsZero: true},
		Background: bgBrush,
		Children: []Widget{
			ComboBox{
				AssignTo:     &s.MethodCB,
				Model:        s.HttpMethods,
				CurrentIndex: 0,
			},
			LineEdit{
				AssignTo: &s.UrlLE,
				Text:     "https://httpbin.org/get",
			},
			PushButton{
				AssignTo:  &s.SendBtn,
				Text:      "Send",
				OnClicked: s.handleSendClicked,
			},
		},
	}
}

func (s *AppState) buildRequestTabs() Widget {
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	return TabWidget{
		Pages: []TabPage{
			{
				Title:      "Headers",
				Layout:     VBox{MarginsZero: true},
				Background: bgBrush,
				Children: []Widget{
					TextEdit{
						AssignTo: &s.HeadersTE,
						Text:     "Content-Type: application/json\n",
						VScroll:  true,
					},
				},
			},
			{
				Title:      "Body",
				Layout:     VBox{MarginsZero: true},
				Background: bgBrush,
				Children: []Widget{
					TextEdit{
						AssignTo: &s.BodyTE,
						Text:     "{\n  \"key\": \"value\"\n}",
						VScroll:  true,
					},
				},
			},
		},
	}
}

func (s *AppState) buildResponseSection() Widget {
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	return Composite{
		Layout:     VBox{MarginsZero: true},
		Background: bgBrush,
		Children: []Widget{
			Label{
				AssignTo: &s.StatusLbl,
				Text:     "Status: N/A",
			},
			TextEdit{
				AssignTo: &s.RespBodyTE,
				ReadOnly: true,
				VScroll:  true,
			},
		},
	}
}

func (s *AppState) handleSendClicked() {
	s.MainWindow.Synchronize(func() {
		s.SendBtn.SetEnabled(false)
		s.StatusLbl.SetText("Status: Sending...")
		s.RespBodyTE.SetText("")
	})

	go func() {
		opts := RequestOpts{
			Method:  s.MethodCB.Text(),
			URL:     s.UrlLE.Text(),
			Headers: s.HeadersTE.Text(),
			Body:    s.BodyTE.Text(),
		}

		res := performRequest(opts)

		s.MainWindow.Synchronize(func() {
			if res.Error != nil {
				s.StatusLbl.SetText("Error")
				s.RespBodyTE.SetText(res.Error.Error())
			} else {
				statusText := fmt.Sprintf("Status: %d %s | Time: %v", res.StatusCode, res.StatusText, res.Duration)
				s.StatusLbl.SetText(statusText)
				s.RespBodyTE.SetText(res.Body)
			}
			s.SendBtn.SetEnabled(true)
		})
	}()
}
