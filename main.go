package main

import (
	"fmt"
	"strings"
	"time"

	html2md "github.com/JohannesKaufmann/html-to-markdown"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	RespBodyTE           *widget.Entry
	RespPreviewRT        *widget.RichText
	EndpointsTree        *widget.Tree
	RightSidebar         *fyne.Container
	RightSidebarVisible  bool
	MainSplit            *container.Split

	Workspace           *Workspace
	ActiveProject       *Project
	ActiveNodeUID       string
	IsUpdatingUI        bool

	HttpMethods []string
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

	ws, err := LoadWorkspace()
	if err != nil {
		dialog.ShowError(fmt.Errorf("could not load workspace: %v", err), w)
		// Fallback to empty if critical fail
		ws = &Workspace{Projects: []Project{{ID: "default", Name: "Default Project"}}}
	}

	state := &AppState{
		App:           a,
		Window:        w,
		Workspace:     ws,
		ActiveProject: &ws.Projects[0],
		ActiveNodeUID: "",
		HttpMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
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

					for _, ep := range eps {
						uid := fmt.Sprintf("node_%d", time.Now().UnixNano())
						time.Sleep(1 * time.Nanosecond)
						node := &Node{
							UID:      uid,
							Name:     ep.Path,
							IsFolder: false,
							Endpoint: &ep,
						}
						s.ActiveProject.Nodes[uid] = node
						s.ActiveProject.RootNodes = append(s.ActiveProject.RootNodes, uid)
					}
					
					SaveWorkspace(s.Workspace)
					if s.EndpointsTree != nil {
						s.EndpointsTree.Refresh()
					}
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
	s.buildRightSidebar()

	s.MainSplit = container.NewHSplit(leftNav, rightPane)
	s.MainSplit.Offset = 0.25 // 25% width for left nav
	
	s.RightSidebarVisible = false
	
	// Post-init load
	s.loadActiveProject()
	
	return s.MainSplit
}

func (s *AppState) loadActiveProject() {
	if s.EndpointsTree == nil || s.VarsTE == nil || s.UrlLE == nil {
		return // Prevent panic during UI construction phase
	}

	s.IsUpdatingUI = true
	s.ActiveNodeUID = ""

	s.EndpointsTree.Refresh()
	
	s.VarsTE.SetText(s.ActiveProject.Variables)
	
	s.UrlLE.SetText("")
	s.HeadersTE.SetText("")
	s.BodyTE.SetText("")
	s.IsUpdatingUI = false
}

func (s *AppState) buildLeftNav() fyne.CanvasObject {
	var projectSelect *widget.Select

	updateSelectOptions := func() {
		projectNames := []string{}
		for _, p := range s.Workspace.Projects {
			projectNames = append(projectNames, p.Name)
		}
		projectSelect.Options = projectNames
		projectSelect.SetSelected(s.ActiveProject.Name)
	}

	projectSelect = widget.NewSelect(nil, func(selected string) {
		for i := range s.Workspace.Projects {
			if s.Workspace.Projects[i].Name == selected {
				s.ActiveProject = &s.Workspace.Projects[i]
				s.loadActiveProject()
				return
			}
		}
	})
	updateSelectOptions()

	newBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		entry := widget.NewEntry()
		dialog.ShowCustomConfirm("New Project", "Create", "Cancel", entry, func(create bool) {
			if create && entry.Text != "" {
				newProj := Project{
					ID:        fmt.Sprintf("proj_%d", time.Now().UnixNano()),
					Name:      entry.Text,
					Nodes:     make(map[string]*Node),
					RootNodes: []string{},
					Variables: "",
				}
				s.Workspace.Projects = append(s.Workspace.Projects, newProj)
				s.ActiveProject = &s.Workspace.Projects[len(s.Workspace.Projects)-1]
				SaveWorkspace(s.Workspace)
				updateSelectOptions()
			}
		}, s.Window)
	})

	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if len(s.Workspace.Projects) <= 1 {
			dialog.ShowInformation("Cannot Delete", "You must have at least one project in the workspace.", s.Window)
			return
		}
		dialog.ShowConfirm("Delete Project", fmt.Sprintf("Delete project '%s'?", s.ActiveProject.Name), func(del bool) {
			if del {
				for i, p := range s.Workspace.Projects {
					if p.ID == s.ActiveProject.ID {
						s.Workspace.Projects = append(s.Workspace.Projects[:i], s.Workspace.Projects[i+1:]...)
						break
					}
				}
				s.ActiveProject = &s.Workspace.Projects[0]
				SaveWorkspace(s.Workspace)
				updateSelectOptions()
			}
		}, s.Window)
	})

	newFolderBtn := widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
		entry := widget.NewEntry()
		dialog.ShowCustomConfirm("New Folder", "Create", "Cancel", entry, func(create bool) {
			if create && entry.Text != "" {
				uid := fmt.Sprintf("folder_%d", time.Now().UnixNano())
				node := &Node{
					UID:      uid,
					Name:     entry.Text,
					IsFolder: true,
				}
				
				// Insert either at root or inside current folder if selected
				if s.ActiveNodeUID != "" && s.ActiveProject.Nodes[s.ActiveNodeUID] != nil && s.ActiveProject.Nodes[s.ActiveNodeUID].IsFolder {
					node.ParentUID = s.ActiveNodeUID
					s.ActiveProject.Nodes[uid] = node
					
					// We need to keep track of folder children, but for simplicity let's scan or we need a Children []string in Node.
					// Let's add children to Node or we can just filter by ParentUID. Scanning by ParentUID is easier.
					// Save Workspace and refresh tree.
				} else {
					s.ActiveProject.RootNodes = append(s.ActiveProject.RootNodes, uid)
					s.ActiveProject.Nodes[uid] = node
				}
				
				SaveWorkspace(s.Workspace)
				s.EndpointsTree.Refresh()
			}
		}, s.Window)
	})

	newReqBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		uid := fmt.Sprintf("req_%d", time.Now().UnixNano())
		newEp := Endpoint{
			Method:  "GET",
			Path:    "/new-request",
			BaseURL: "http://localhost:8080",
			Headers: "Content-Type: application/json",
		}
		node := &Node{
			UID:      uid,
			Name:     "/new-request",
			IsFolder: false,
			Endpoint: &newEp,
		}

		if s.ActiveNodeUID != "" && s.ActiveProject.Nodes[s.ActiveNodeUID] != nil {
			if s.ActiveProject.Nodes[s.ActiveNodeUID].IsFolder {
				node.ParentUID = s.ActiveNodeUID
			} else {
				node.ParentUID = s.ActiveProject.Nodes[s.ActiveNodeUID].ParentUID
			}
		}
		
		s.ActiveProject.Nodes[uid] = node
		if node.ParentUID == "" {
			s.ActiveProject.RootNodes = append(s.ActiveProject.RootNodes, uid)
		}
		SaveWorkspace(s.Workspace)
		s.EndpointsTree.Refresh()
	})

	title := widget.NewLabel("Projects & APIs")
	title.TextStyle = fyne.TextStyle{Bold: true}
	
	reqRow := container.NewBorder(nil, nil, nil, container.NewHBox(newFolderBtn, newReqBtn), title)

	s.EndpointsTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return s.ActiveProject.RootNodes
			}
			var children []string
			for uid, n := range s.ActiveProject.Nodes {
				if n.ParentUID == id {
					children = append(children, uid)
				}
			}
			return children
		},
		func(id widget.TreeNodeID) bool {
			if id == "" { return true }
			if n, ok := s.ActiveProject.Nodes[id]; ok {
				return n.IsFolder
			}
			return false
		},
		func(branch bool) fyne.CanvasObject {
			var icon fyne.Resource
			if branch {
				icon = theme.FolderIcon()
			} else {
				icon = theme.DocumentIcon()
			}
			
			btn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
			lbl := widget.NewLabel("Loading...")
			return container.NewBorder(nil, nil, widget.NewIcon(icon), btn, lbl)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			node, ok := s.ActiveProject.Nodes[id]
			if !ok { return }

			border := o.(*fyne.Container)
			
			var lbl *widget.Label
			var btnBox *widget.Button
			for _, obj := range border.Objects {
				switch v := obj.(type) {
				case *widget.Label:
					lbl = v
				case *widget.Button:
					btnBox = v
				}
			}
			
			if lbl == nil || btnBox == nil { return }

			if branch {
				lbl.SetText(node.Name)
			} else {
				lbl.SetText(fmt.Sprintf("%-7s %s", node.Endpoint.Method, node.Endpoint.Path))
			}

			// Capture id for click action
			btnBox.OnTapped = func() {
				dialog.ShowConfirm("Delete Node", fmt.Sprintf("Delete '%s'?", node.Name), func(del bool) {
					if del {
						if node.ParentUID == "" {
							for i, ruid := range s.ActiveProject.RootNodes {
								if ruid == id {
									s.ActiveProject.RootNodes = append(s.ActiveProject.RootNodes[:i], s.ActiveProject.RootNodes[i+1:]...)
									break
								}
							}
						}
						
						// Recursive delete would be needed, but flat map allows easy cleanup by tracking children.
						// simplified: just delete the node and let orphans float or clean them here recursively
						delete(s.ActiveProject.Nodes, id)
						
						if s.ActiveNodeUID == id {
							s.ActiveNodeUID = ""
							s.loadActiveProject()
						}
						SaveWorkspace(s.Workspace)
						s.EndpointsTree.Refresh()
					}
				}, s.Window)
			}
		},
	)

	s.EndpointsTree.OnSelected = func(id widget.TreeNodeID) {
		node, ok := s.ActiveProject.Nodes[id]
		if !ok { return }
		
		s.IsUpdatingUI = true
		s.ActiveNodeUID = id
		
		if node.IsFolder || node.Endpoint == nil {
			s.UrlLE.SetText("")
			s.HeadersTE.SetText("")
			s.BodyTE.SetText("")
			s.IsUpdatingUI = false
			return
		}
		
		ep := node.Endpoint

		s.MethodCB.SetSelected(ep.Method)

		fullURL := ep.Path
		if ep.BaseURL != "" {
			fullURL = ep.BaseURL + ep.Path
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
		s.IsUpdatingUI = false
	}

	projectRow := container.NewBorder(nil, nil, nil, container.NewHBox(newBtn, delBtn), projectSelect)
	topBox := container.NewVBox(projectRow, reqRow)
	return container.NewBorder(topBox, nil, nil, nil, s.EndpointsTree)
}

func (s *AppState) buildRightPane() fyne.CanvasObject {
	topBar := s.buildTopBar()

	s.HeadersTE = widget.NewMultiLineEntry()
	s.HeadersTE.SetText("Content-Type: application/json\n")
	s.HeadersTE.OnChanged = func(val string) {
		if !s.IsUpdatingUI && s.ActiveNodeUID != "" {
			if node, ok := s.ActiveProject.Nodes[s.ActiveNodeUID]; ok && node.Endpoint != nil {
				node.Endpoint.Headers = val
				SaveWorkspace(s.Workspace)
			}
		}
	}

	s.BodyTE = widget.NewMultiLineEntry()
	s.BodyTE.SetText("{\n  \"key\": \"value\"\n}")
	s.BodyTE.OnChanged = func(val string) {
		if !s.IsUpdatingUI && s.ActiveNodeUID != "" {
			if node, ok := s.ActiveProject.Nodes[s.ActiveNodeUID]; ok && node.Endpoint != nil {
				node.Endpoint.Body = val
				SaveWorkspace(s.Workspace)
			}
		}
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Headers", s.HeadersTE),
		container.NewTabItem("Body", s.BodyTE),
	)

	s.StatusLbl = widget.NewLabel("Status: N/A")
	
	s.RespBodyTE = widget.NewMultiLineEntry()
	
	s.RespPreviewRT = widget.NewRichTextFromMarkdown("Preview not available.")
	previewScroll := container.NewScroll(s.RespPreviewRT)

	respTabs := container.NewAppTabs(
		container.NewTabItem("Raw", s.RespBodyTE),
		container.NewTabItem("Preview", previewScroll),
	)

	respSection := container.NewBorder(s.StatusLbl, nil, nil, nil, respTabs)

	split := container.NewVSplit(tabs, respSection)
	split.Offset = 0.5

	return container.NewBorder(topBar, nil, nil, nil, split)
}

func (s *AppState) buildRightSidebar() fyne.CanvasObject {
	s.VarsTE = widget.NewMultiLineEntry()
	s.VarsTE.SetText(s.ActiveProject.Variables)
	s.VarsTE.OnChanged = func(val string) {
		s.ActiveProject.Variables = val
		SaveWorkspace(s.Workspace)
	}

	title := widget.NewLabel("Global Variables")
	title.TextStyle = fyne.TextStyle{Bold: true}
	
	info := widget.NewLabel("Use {{KEY}} in requests")
	info.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewBorder(container.NewVBox(title, info), nil, nil, nil, s.VarsTE)
	s.RightSidebar = container.NewStack(content)
	
	// We don't hide it so it renders perfectly when added to split
	return s.RightSidebar
}

func (s *AppState) buildTopBar() fyne.CanvasObject {
	s.MethodCB = widget.NewSelect(s.HttpMethods, func(val string) {
		if !s.IsUpdatingUI && s.ActiveNodeUID != "" {
			if node, ok := s.ActiveProject.Nodes[s.ActiveNodeUID]; ok && node.Endpoint != nil {
				node.Endpoint.Method = val
				s.EndpointsTree.Refresh()
				SaveWorkspace(s.Workspace)
			}
		}
	})
	s.MethodCB.SetSelected("GET")

	s.UrlLE = widget.NewEntry()
	s.UrlLE.SetText("https://httpbin.org/get")
	s.UrlLE.OnChanged = func(val string) {
		if !s.IsUpdatingUI && s.ActiveNodeUID != "" {
			if node, ok := s.ActiveProject.Nodes[s.ActiveNodeUID]; ok && node.Endpoint != nil {
				node.Endpoint.Path = val
				node.Endpoint.BaseURL = ""
				s.EndpointsTree.Refresh()
				SaveWorkspace(s.Workspace)
			}
		}
	}

	s.SendBtn = widget.NewButton("Send", s.handleSendClicked)
	s.SendBtn.Importance = widget.HighImportance

	toggleBtn := widget.NewButtonWithIcon("", theme.MenuExpandIcon(), func() {
		if s.RightSidebarVisible {
			s.RightSidebarVisible = false
			s.Window.SetContent(s.MainSplit)
		} else {
			s.RightSidebarVisible = true
			split := container.NewHSplit(s.MainSplit, s.RightSidebar)
			split.Offset = 0.75 // 75% for main content, 25% for sidebar
			s.Window.SetContent(split)
		}
	})

	rightBox := container.NewHBox(s.SendBtn, toggleBtn)

	return container.NewBorder(nil, nil, s.MethodCB, rightBox, s.UrlLE)
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
				s.RespPreviewRT.ParseMarkdown("")
			} else {
				statusText := fmt.Sprintf("Status: %d %s | Time: %v | Type: %s", res.StatusCode, res.StatusText, res.Duration, res.ContentType)
				s.StatusLbl.SetText(statusText)
				s.RespBodyTE.SetText(res.Body)

				// Handle HTML to Markdown substitution
				if strings.Contains(strings.ToLower(res.ContentType), "text/html") {
					converter := html2md.NewConverter("", true, nil)
					markdown, err := converter.ConvertString(res.Body)
					if err == nil {
						s.RespPreviewRT.ParseMarkdown(markdown)
					} else {
						s.RespPreviewRT.ParseMarkdown("*Failed to render HTML preview*")
					}
				} else {
					s.RespPreviewRT.ParseMarkdown("*Preview not available for this content type.*")
				}
			}
			s.SendBtn.Enable()
		})
	}()
}
