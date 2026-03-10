package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// AppState manages the UI widgets
type AppState struct {
	App           fyne.App
	Window        fyne.Window
	MethodCB      *widget.Select
	UrlLE         *widget.Entry
	SendBtn       *widget.Button
	HeadersTE     *widget.Entry
	BodyTE        *widget.Entry
	VarsTE        *widget.Entry
	StatusLbl     *widget.Label
	RespBodyTE    *widget.Entry
	EndpointsList *widget.List

	Endpoints       []Endpoint
	EndpointStrings []string
	HttpMethods     []string
}

func main() {
	a := app.New()
	w := a.NewWindow("Postlang")
	w.Resize(fyne.NewSize(900, 600))

	logo, err := fyne.LoadResourceFromPath("logo.png")
	if err == nil {
		w.SetIcon(logo)
		a.SetIcon(logo)
	}

	state := &AppState{
		App:         a,
		Window:      w,
		HttpMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
	}

	w.SetMainMenu(state.buildMenuBar())
	w.SetContent(state.buildUI())

	w.ShowAndRun()
}

func (s *AppState) buildMenuBar() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Import OpenAPI Spec...", func() {
				fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, s.Window)
						return
					}
					if reader == nil {
						return
					}
					defer reader.Close()

					eps, err := parseOpenAPI(reader)
					if err != nil {
						dialog.ShowError(fmt.Errorf("error Loading OpenAPI: %v", err), s.Window)
						return
					}

					s.Endpoints = eps
					s.EndpointStrings = make([]string, len(s.Endpoints))
					for i, e := range s.Endpoints {
						s.EndpointStrings[i] = e.DisplayName()
					}

					s.EndpointsList.Refresh()
				}, s.Window)

				// Fyne file filters
				// fd.SetFilter(storage.NewExtensionFileFilter([]string{".json", ".yaml", ".yml"}))
				
				fd.Show()
			}),
		),
	)
}

func (s *AppState) buildUI() fyne.CanvasObject {
	leftNav := s.buildLeftNav()
	rightPane := s.buildRightPane()

	split := container.NewHSplit(leftNav, rightPane)
	split.Offset = 0.3 // 30% width for left nav
	return split
}

func (s *AppState) buildLeftNav() fyne.CanvasObject {
	title := widget.NewLabel("API Endpoints (Import from File)")
	title.TextStyle = fyne.TextStyle{Bold: true}

	s.EndpointsList = widget.NewList(
		func() int { return len(s.EndpointStrings) },
		func() fyne.CanvasObject { return widget.NewLabel("Method /path/to/endpoint") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(s.EndpointStrings[i])
		},
	)

	s.EndpointsList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || int(id) >= len(s.Endpoints) {
			return
		}
		ep := s.Endpoints[id]

		s.MethodCB.SetSelected(ep.Method)

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

	return container.NewBorder(title, nil, nil, nil, s.EndpointsList)
}

func (s *AppState) buildRightPane() fyne.CanvasObject {
	topBar := s.buildTopBar()

	s.HeadersTE = widget.NewMultiLineEntry()
	s.HeadersTE.SetText("Content-Type: application/json\n")

	s.BodyTE = widget.NewMultiLineEntry()
	s.BodyTE.SetText("{\n  \"key\": \"value\"\n}")

	s.VarsTE = widget.NewMultiLineEntry()
	s.VarsTE.SetText("BASE_URL=https://httpbin.org\nTOKEN=my-secret-token")

	tabs := container.NewAppTabs(
		container.NewTabItem("Headers", s.HeadersTE),
		container.NewTabItem("Body", s.BodyTE),
		container.NewTabItem("Variables", s.VarsTE),
	)

	s.StatusLbl = widget.NewLabel("Status: N/A")
	
	s.RespBodyTE = widget.NewMultiLineEntry()
	// To make it readonly, we disable it, or just use NewMultiLineEntry in read-only mode if Fyne allows. Fyne Entries don't have a simple ReadOnly bool in v2, but we can Disable it, which changes appearance. Let's just leave it editable but not updateable user-side since Postman lets you copy text. So we leave it standard.

	respSection := container.NewBorder(s.StatusLbl, nil, nil, nil, s.RespBodyTE)

	split := container.NewVSplit(tabs, respSection)
	split.Offset = 0.5

	return container.NewBorder(topBar, nil, nil, nil, split)
}

func (s *AppState) buildTopBar() fyne.CanvasObject {
	s.MethodCB = widget.NewSelect(s.HttpMethods, nil)
	s.MethodCB.SetSelected("GET")

	s.UrlLE = widget.NewEntry()
	s.UrlLE.SetText("https://httpbin.org/get")

	s.SendBtn = widget.NewButton("Send", s.handleSendClicked)
	s.SendBtn.Importance = widget.HighImportance

	return container.NewBorder(nil, nil, s.MethodCB, s.SendBtn, s.UrlLE)
}

func (s *AppState) handleSendClicked() {
	s.SendBtn.Disable()
	s.StatusLbl.SetText("Status: Sending...")
	s.RespBodyTE.SetText("")

	go func() {
		// Parse variables
		vars := make(map[string]string)
		for _, line := range strings.Split(s.VarsTE.Text, "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				vars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		// Substitution function
		substitute := func(in string) string {
			out := in
			for k, v := range vars {
				out = strings.ReplaceAll(out, "{{"+k+"}}", v)
			}
			return out
		}

		opts := RequestOpts{
			Method:  s.MethodCB.Selected,
			URL:     substitute(s.UrlLE.Text),
			Headers: substitute(s.HeadersTE.Text),
			Body:    substitute(s.BodyTE.Text),
		}

		res := performRequest(opts)

		// Schedule UI update on main thread
		// (Though Fyne widgets usually hander bindings safely)
		time.AfterFunc(10*time.Millisecond, func() {
			if res.Error != nil {
				s.StatusLbl.SetText("Error: " + res.Error.Error())
				s.RespBodyTE.SetText("")
			} else {
				statusText := fmt.Sprintf("Status: %d %s | Time: %v", res.StatusCode, res.StatusText, res.Duration)
				s.StatusLbl.SetText(statusText)
				s.RespBodyTE.SetText(res.Body)
			}
			s.SendBtn.Enable()
		})
	}()
}
