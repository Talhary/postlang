package main

import (
	"fmt"
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var methodCB *walk.ComboBox
	var urlLE *walk.LineEdit
	var sendBtn *walk.PushButton
	var reqHeadersTE *walk.TextEdit
	var reqBodyTE *walk.TextEdit
	var respStatusLbl *walk.Label
	var respBodyTE *walk.TextEdit
	var epLB *walk.ListBox

	endpoints := []Endpoint{}
	endpointStrings := []string{}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	// Define Dark Theme Colors
	bgBrush := SolidColorBrush{Color: walk.RGB(40, 44, 52)}
	// Note: Walk TextColor property isn't available declaratively on all widgets, 
	// but we try to set it where supported via direct properties if needed.

	if _, err := (MainWindow{
		AssignTo:   &mw,
		Title:      "Postlang",
		MinSize:    Size{Width: 900, Height: 600},
		Background: bgBrush,
		Font:       Font{Family: "Segoe UI", PointSize: 10},
		Layout:     HBox{MarginsZero: true},
		MenuItems: []MenuItem{
			Menu{
				Text: "&File",
				Items: []MenuItem{
					Action{
						Text: "Import OpenAPI Spec...",
						OnTriggered: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Import OpenAPI Spec"
							dlg.Filter = "OpenAPI Files (*.json;*.yaml;*.yml)|*.json;*.yaml;*.yml|All Files (*.*)|*.*"

							if ok, err := dlg.ShowOpen(mw); err != nil {
								return
							} else if !ok {
								return
							}

							eps, err := parseOpenAPI(dlg.FilePath)
							if err != nil {
								walk.MsgBox(mw, "Error Loading OpenAPI", err.Error(), walk.MsgBoxIconError)
								return
							}

							endpoints = eps
							endpointStrings = make([]string, len(endpoints))
							for i, e := range endpoints {
								endpointStrings[i] = e.DisplayName()
							}
							
							epLB.SetModel(endpointStrings)
						},
					},
				},
			},
		},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					// Left Navigation Pane
					Composite{
						Layout: VBox{MarginsZero: true},
						MinSize: Size{Width: 250, Height: 0},
						Children: []Widget{
							Label{
								Text: "API Endpoints (Import from File)",
								TextColor: walk.RGB(220, 220, 220),
								Background: bgBrush,
							},
							ListBox{
								AssignTo: &epLB,
								Model:    endpointStrings,
								OnCurrentIndexChanged: func() {
									idx := epLB.CurrentIndex()
									if idx >= 0 && idx < len(endpoints) {
										ep := endpoints[idx]
										methodCB.SetText(ep.Method)
										// We just display the path. For a real URL, the user would prepend the base URL.
										urlLE.SetText("http://localhost:8080" + ep.Path)
									}
								},
							},
						},
					},
					
					// Right Pane
					Composite{
						Layout: VBox{},
						Background: bgBrush,
						Children: []Widget{
							// Top Bar
							Composite{
								Layout: HBox{MarginsZero: true},
								Background: bgBrush,
								Children: []Widget{
									ComboBox{
										AssignTo:      &methodCB,
										Model:         methods,
										CurrentIndex:  0,
									},
									LineEdit{
										AssignTo: &urlLE,
										Text:     "https://httpbin.org/get",
									},
									PushButton{
										AssignTo: &sendBtn,
										Text:     "Send",
										OnClicked: func() {
											mw.Synchronize(func() {
												sendBtn.SetEnabled(false)
												respStatusLbl.SetText("Status: Sending...")
												respBodyTE.SetText("")
											})

											// Execute the request asynchronously
											go func() {
												opts := RequestOpts{
													Method:  methodCB.Text(),
													URL:     urlLE.Text(),
													Headers: reqHeadersTE.Text(),
													Body:    reqBodyTE.Text(),
												}

												res := performRequest(opts)

												mw.Synchronize(func() {
													if res.Error != nil {
														respStatusLbl.SetText("Error")
														respBodyTE.SetText(res.Error.Error())
													} else {
														statusText := fmt.Sprintf("Status: %d %s | Time: %v", res.StatusCode, res.StatusText, res.Duration)
														respStatusLbl.SetText(statusText)
														respBodyTE.SetText(res.Body)
													}
													sendBtn.SetEnabled(true)
												})
											}()
										},
									},
								},
							},
							// Request / Response Splitter
							VSplitter{
								Children: []Widget{
									// Request Section
									TabWidget{
										Pages: []TabPage{
											{
												Title: "Headers",
												Layout: VBox{MarginsZero: true},
												Background: bgBrush,
												Children: []Widget{
													TextEdit{
														AssignTo: &reqHeadersTE,
														Text:     "Content-Type: application/json\n",
														VScroll:  true,
													},
												},
											},
											{
												Title: "Body",
												Layout: VBox{MarginsZero: true},
												Background: bgBrush,
												Children: []Widget{
													TextEdit{
														AssignTo: &reqBodyTE,
														Text:     "{\n  \"key\": \"value\"\n}",
														VScroll:  true,
													},
												},
											},
										},
									},
									// Response Section
									Composite{
										Layout: VBox{MarginsZero: true},
										Background: bgBrush,
										Children: []Widget{
											Label{
												AssignTo: &respStatusLbl,
												Text:     "Status: N/A",
												TextColor: walk.RGB(220, 220, 220),
												Background: bgBrush,
											},
											TextEdit{
												AssignTo: &respBodyTE,
												ReadOnly: true,
												VScroll:  true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
