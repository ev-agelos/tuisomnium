package main

import (
	"log"
	"slices"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var mode, view int

const (
	normal int = iota
	insert
)
const (
	// tab-able views
	requestsView int = iota
	httpMethodView
	urlView
	sendView
	bodyTypeView
	authTypeView
	queryView
	headersView
	historyView
	responseView
	responseHeadersView
	responseCookiesView
	responseLogs
	// non tab-able views
	methodOptionsView
	bodyOptionsView
	authOptionsView
	bodyContentView
	authContentView
	queryContentView
	headersContentView
	historyOptionsView
)

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case response:
		res := saveResponse(m, msg)
		m.requests.items[m.requests.cursor].Responses = append(m.requests.items[m.requests.cursor].Responses, res)
		// now update the UI
		m.resBody.SetContent(msg.body)
		m.resLogs = msg.traceLogs
		setTableRows(&m.resHeaders, msg.headers)
		setTableRows(&m.resCookies, msg.cookies)
	case errMsg:
		log.Fatal(m.err.Error())
		return m, tea.Quit
	case tea.KeyMsg:
		key := msg.String()
		switch mode {
		case normal:
			switch key {
			case "n":
				switch view {
				case requestsView:
					tempName := makeInputField("New Request", "")
					m.names = append(m.names, tempName)
					m.requests.cursor = len(m.names) - 1
					req := &DBRequest{}
					for _, option := range m.body.options {
						req.Body.Types = append(req.Body.Types, &NameValue{option.type_, ""})
					}
					m.requests.items = append(m.requests.items, req)
					m.names[m.requests.cursor].Focus()
					mode = insert
				case queryView:
					name, value := makeInputField("", "name"), makeInputField("", "value")
					m.queryParams.fields = append(m.queryParams.fields, name, value)
					m.queryParams.cursor = len(m.queryParams.fields) - 2 // name
					name.Focus()
					view = queryContentView
					mode = insert
				case headersView:
					name, value := makeInputField("", "header"), makeInputField("", "value")
					m.headers.fields = append(m.headers.fields, name, value)
					m.headers.cursor = len(m.headers.fields) - 2 // name
					name.Focus()
					view = headersContentView
					mode = insert
				}
			case "d":
				switch view {
				case requestsView:
					if len(m.names) == 0 {
						break
					}
					var err error
					if err = m.db.DeleteRequest(m.requests.items[m.requests.cursor]); err != nil {
						// TODO: handle error instead of quitting
						return m, tea.Quit
					}
					m.names = slices.Delete(m.names, m.requests.cursor, m.requests.cursor+1)
					m.requests.items = slices.Delete(m.requests.items, m.requests.cursor, m.requests.cursor+1)
					m.requests.cursor = min(m.requests.cursor, len(m.names)-1)
					m.requests.cursor = max(m.requests.cursor, 0)
					if len(m.requests.items) > 0 {
						setUIRequest(&m, m.requests.items[m.requests.cursor])
					} else {
						m.method.cursor = 0
						m.url.SetValue("")
					}
				}
			case "enter":
				switch view {
				case requestsView:
					cmd = m.names[m.requests.cursor].Focus()
					mode = insert
					return m, cmd
				case httpMethodView:
					view = methodOptionsView
				case methodOptionsView:
					m.requests.items[m.requests.cursor].Method = m.method.options[m.method.cursor]
					saveRequest(m)
					view = httpMethodView
				case urlView:
					cmd = m.url.Focus()
					mode = insert
					return m, cmd
				case sendView:
					return m, sendRequest(m)
				case bodyTypeView:
					view = bodyOptionsView
				case bodyOptionsView:
					m.body.field.SetValue(m.body.options[m.body.cursor].value)
					saveRequest(m)
					view = bodyTypeView
				case authTypeView:
					view = authOptionsView
				case authOptionsView:
					m.requests.items[m.requests.cursor].Auth.Selected = m.auth.cursor
					saveRequest(m)
					view = authTypeView
				case queryContentView:
					cmd = m.queryParams.fields[m.queryParams.cursor].Focus()
					mode = insert
					return m, cmd
				case headersView:
					view = headersContentView
				case headersContentView:
					cmd = m.headers.fields[m.headers.cursor].Focus()
					mode = insert
					return m, cmd
				case authContentView:
					authOption := m.auth.options[m.auth.cursor]
					cmd = authOption.fields[authOption.cursor].Focus()
					mode = insert
					return m, cmd
				case bodyContentView:
					if m.body.cursor == 0 { // No body option
						break
					}
					cmd = m.body.field.Focus()
					mode = insert
					return m, cmd
				case responseView:
					mode = insert // lame way to "focus" the response viewport
				case responseHeadersView:
					m.resHeaders.Focus()
					mode = insert
				case responseCookiesView:
					m.resCookies.Focus()
					mode = insert
				case historyView:
					view = historyOptionsView
				case historyOptionsView:
					m.resBody.SetContent(m.requests.items[m.requests.cursor].Responses[m.history.cursor].Body)
					setTableRows(&m.resHeaders, m.requests.items[m.requests.cursor].Responses[m.history.cursor].Headers)
					setTableRows(&m.resCookies, m.requests.items[m.requests.cursor].Responses[m.history.cursor].Cookies)
					view = historyView
				}
			case "j":
				switch view {
				case requestsView:
					m.requests.down()
					setUIRequest(&m, m.requests.items[m.requests.cursor])
				case httpMethodView:
					view = bodyTypeView + m.requestPaginator.Page
				case methodOptionsView:
					m.method.down()
				case urlView, sendView:
					view = bodyTypeView + m.requestPaginator.Page
				case bodyTypeView:
					if m.body.cursor == 0 {
						view = historyView
					} else {
						view = bodyContentView
					}
				case bodyOptionsView:
					m.body.down()
				case bodyContentView:
					view = historyView
				case authTypeView:
					if m.auth.cursor == 0 {
						view = historyView
					} else {
						view = authContentView
					}
				case authOptionsView:
					m.auth.down()
				case authContentView:
					t := m.auth.options[m.auth.cursor]
					if !t.down() {
						view = historyView
					}
				case queryView:
					view = queryContentView
				case headersView:
					view = headersContentView
				case headersContentView:
					if !m.headers.down() {
						view = historyView
					}
				case queryContentView:
					if !m.queryParams.down() {
						view = historyView
					}
				case historyView:
					view = responseView + m.responsePaginator.Page
				case historyOptionsView:
					m.history.cursor = min(m.history.cursor+1, len(m.requests.items[m.requests.cursor].Responses)-1)
				}
			case "k":
				switch view {
				case requestsView:
					m.requests.up()
					setUIRequest(&m, m.requests.items[m.requests.cursor])
				case methodOptionsView:
					m.method.up()
				case bodyTypeView:
					view = httpMethodView
				case bodyOptionsView:
					m.body.up()
				case authTypeView:
					view = httpMethodView
				case authOptionsView:
					m.auth.up()
				case authContentView:
					t := m.auth.options[m.auth.cursor]
					if !t.up() {
						view = authTypeView
					}
				case queryView, headersView:
					view = httpMethodView
				case bodyContentView:
					view = bodyTypeView + m.requestPaginator.Page
				case queryContentView:
					if !m.queryParams.up() {
						view = queryView
					}
				case headersContentView:
					if !m.headers.up() {
						view = headersView
					}
				case historyView:
					switch m.requestPaginator.Page {
					case 0:
						if m.body.cursor != 0 {
							view = bodyContentView
						} else {
							view = bodyTypeView
						}
					case 1:
						if m.auth.cursor != 0 {
							view = authContentView
						} else {
							view = authTypeView
						}
					case 2:
						view = queryContentView
					case 3:
						view = headersContentView
					}
				case historyOptionsView:
					m.history.cursor = max(m.history.cursor-1, 0)
				case responseView, responseHeadersView, responseCookiesView, responseLogs:
					view = historyView
				}
			case "tab", "l":
				switch view {
				case requestsView, httpMethodView, urlView:
					view++
				case bodyTypeView, authTypeView, queryView:
					view++
					m.requestPaginator.NextPage()
				case queryContentView:
					m.queryParams.cursor = min(m.queryParams.cursor+1, len(m.queryParams.fields)-1)
				case headersContentView:
					m.headers.cursor = min(m.headers.cursor+1, len(m.headers.fields)-1)
				case responseView, responseHeadersView, responseCookiesView:
					m.responsePaginator.NextPage()
					view++
				}
			case "shift+tab", "h":
				switch view {
				case httpMethodView, urlView, sendView:
					view--
				case bodyTypeView, responseView:
					view = requestsView
				case authTypeView, queryView, headersView:
					m.requestPaginator.PrevPage()
					view--
				case queryContentView:
					m.queryParams.cursor = max(m.queryParams.cursor-1, 0)
				case headersContentView:
					m.headers.cursor = max(m.headers.cursor-1, 0)
				case responseHeadersView, responseCookiesView, responseLogs:
					m.responsePaginator.PrevPage()
					view--
				}
			case "esc":
				switch view {
				case methodOptionsView:
					view = httpMethodView
				case bodyOptionsView:
					view = bodyTypeView
				case authOptionsView:
					view = authTypeView
				case queryContentView:
					view = queryView
				case headersContentView:
					view = headersView
				case authContentView:
					view = authTypeView
				case historyOptionsView:
					view = historyView
				case responseHeadersView:
					m.resHeaders.Blur()
				case responseCookiesView:
					m.resCookies.Blur()
				}
			case "q":
				return m, tea.Quit
			}
		case insert:
			switch view {
			case requestsView:
				switch key {
				case "esc":
					m.requests.items[m.requests.cursor].Name = m.names[m.requests.cursor].Value()
					m.names[m.requests.cursor].Blur()
					mode = normal
					saveRequest(m)
				}
			case urlView:
				switch key {
				case "esc":
					m.requests.items[m.requests.cursor].Url = m.url.Value()
					m.url.Blur()
					mode = normal
					saveRequest(m)
				}
			case bodyContentView:
				switch key {
				case "esc":
					m.body.options[m.body.cursor].value = m.body.field.Value()
					m.body.field.Blur()
					saveRequest(m)
					mode = normal
				}
			case queryContentView:
				switch key {
				case "esc":
					m.requests.items[m.requests.cursor].Query = nil
					for i := range len(m.queryParams.fields) / 2 {
						name := m.queryParams.fields[i*2].Value()
						value := m.queryParams.fields[i*2+1].Value()
						if len(name) == 0 && len(value) == 0 {
							continue
						}
						m.requests.items[m.requests.cursor].Query = append(m.requests.items[m.requests.cursor].Query, NameValue{name, value})
					}
					m.queryParams.fields[m.queryParams.cursor].Blur()
					saveRequest(m)
					mode = normal
				}
			case headersContentView:
				switch key {
				case "esc":
					m.requests.items[m.requests.cursor].Headers = nil
					for i := range len(m.headers.fields) / 2 {
						name := m.headers.fields[i*2].Value()
						value := m.headers.fields[i*2+1].Value()
						if len(name) == 0 && len(value) == 0 {
							continue
						}
						m.requests.items[m.requests.cursor].Headers = append(m.requests.items[m.requests.cursor].Headers, NameValue{name, value})
					}
					m.headers.fields[m.headers.cursor].Blur()
					saveRequest(m)
					mode = normal
				}
			case authContentView:
				authType := m.auth.options[m.auth.cursor]
				switch key {
				case "esc":
					value := authType.fields[authType.cursor].Value()
					m.requests.items[m.requests.cursor].Auth.Types[m.auth.cursor].Fields[authType.cursor].Value = value
					authType.fields[authType.cursor].Blur()
					saveRequest(m)
					mode = normal
				}
			case responseView, responseHeadersView, responseCookiesView:
				switch key {
				case "esc":
					mode = normal // again lame way to "unfocus" the response viewport
				}
			}
		}
	}

	if len(m.requests.items) == 0 {
		return m, cmd
	}

	switch view {
	case requestsView:
		if m.names[m.requests.cursor].Focused() {
			m.names[m.requests.cursor], cmd = m.names[m.requests.cursor].Update(msg)
		}
	case urlView:
		if m.url.Focused() {
			m.url, cmd = m.url.Update(msg)
		}
	case bodyContentView:
		if m.body.field.Focused() {
			m.body.field, cmd = m.body.field.Update(msg)
		}
	case authContentView:
		if m.auth.cursor == 0 {
			break
		}
		option := m.auth.options[m.auth.cursor]
		field := option.fields[option.cursor]
		if field.Focused() {
			option.fields[option.cursor], cmd = option.fields[option.cursor].Update(msg)
		}
	case queryContentView:
		if m.queryParams.fields[m.queryParams.cursor].Focused() {
			m.queryParams.fields[m.queryParams.cursor], cmd = m.queryParams.fields[m.queryParams.cursor].Update(msg)
		}
	case headersContentView:
		if m.headers.fields[m.headers.cursor].Focused() {
			m.headers.fields[m.headers.cursor], cmd = m.headers.fields[m.headers.cursor].Update(msg)
		}
	case responseView:
		if mode == insert {
			m.resBody, cmd = m.resBody.Update(msg)
		}
	case responseHeadersView:
		if m.resHeaders.Focused() {
			m.resHeaders, cmd = m.resHeaders.Update(msg)
		}
	case responseCookiesView:
		if m.resCookies.Focused() {
			m.resCookies, cmd = m.resCookies.Update(msg)
		}
	}
	return m, cmd
}

func saveRequest(m model) {
	req := m.requests.items[m.requests.cursor]
	if req == nil && req.Name == "" {
		return
	}
	req.Method = m.method.options[m.method.cursor]
	req.Url = m.url.Value()

	for i, option := range m.body.options {
		req.Body.Types[i].Value = option.value
	}
	req.Body.Selected = m.body.cursor

	// Auth
	if len(req.Auth.Types) == 0 {
		for i, option := range m.auth.options {
			req.Auth.Types = append(req.Auth.Types, dbAuthType{option.name, []*NameValue{}})
			for _, field := range option.fields {
				req.Auth.Types[i].Fields = append(req.Auth.Types[i].Fields, &NameValue{field.Placeholder, ""})
			}
		}
	}
	for i, type_ := range m.auth.options {
		for j, field := range type_.fields {
			req.Auth.Types[i].Fields[j].Value = field.Value()
		}

	}
	req.Auth.Selected = m.auth.cursor

	var err error
	if err = m.db.SaveRequest(req); err != nil {
		log.Fatal("Error saving request: ", err)
	}
}

func saveResponse(m model, msg response) DBResponse {
	response := DBResponse{
		0,
		m.requests.items[m.requests.cursor].ID,
		m.requests.items[m.requests.cursor].Method,
		m.requests.items[m.requests.cursor].Url,
		msg.body,
		msg.status,
		msg.headers,
		msg.cookies,
		msg.duration,
		msg.contentLength,
		msg.responseAt,
		msg.traceLogs,
	}
	var err error
	if err = m.db.SaveResponse(&response); err != nil {
		log.Fatal("Error saving response: ", err)
	}
	return response
}

func setTableRows(t *table.Model, pairs []NameValue) {
	rows := []table.Row{}
	for _, pair := range pairs {
		rows = append(rows, table.Row{pair.Name, pair.Value})
	}
	t.SetRows(rows)
}
