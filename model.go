package main

import (
	"log"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type requestList struct {
	items  []*DBRequest
	cursor int
}

func (l *requestList) down() {
	l.cursor = min(l.cursor+1, max(0, len(l.items)-1))
}
func (l *requestList) up() {
	l.cursor = max(l.cursor-1, 0)
}

type model struct {
	db       *Store
	requests requestList

	names []textinput.Model

	method           selectField
	url              textinput.Model
	body             body
	auth             auth
	queryParams      inputFields
	headers          inputFields
	requestPaginator paginator.Model
	requestContents  []string

	history           selectField
	resBody           viewport.Model
	resHeaders        table.Model
	resCookies        table.Model
	resLogs           string
	responsePaginator paginator.Model

	err error
}

func initialModel(store *Store) model {
	url := makeInputField("", "https...")
	m := model{
		db:               store,
		requests:         requestList{},
		method:           makeSelectField("GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"),
		url:              url,
		body:             makeBody(),
		auth:             makeAuth(),
		resBody:          viewport.New(40, 10),
		resLogs:          "",
		requestContents:  []string{"", "", "", ""},
	}

	m = setupUI(m)

	requests, err := store.GetRequests()
	if err != nil {
		log.Fatalf("unable to get saved requests: %v", err)
	}
	if len(requests) == 0 {
		return m
	}

	setUIRequestNames(&m, requests)
	setUIRequest(&m, &requests[0])

	return m
}

func setUIRequestNames(m *model, requests []DBRequest) {
	for _, r := range requests {
		m.requests.items = append(m.requests.items, &r)
		field := makeInputField(r.Name, "")
		m.names = append(m.names, field)
	}
}

func setupUI(m model) model {
	t := paginator.New()
	t.KeyMap.NextPage.SetEnabled(false)
	t.KeyMap.PrevPage.SetEnabled(false)
	t.Type = paginator.Dots
	t.PerPage = 1
	t.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Background(primaryColor).Render("•")
	t.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Background(primaryColor).Render("•")
	t.SetTotalPages(len(m.requestContents))
	m.requestPaginator = t

	p := paginator.New()
	p.KeyMap.NextPage.SetEnabled(false)
	p.KeyMap.PrevPage.SetEnabled(false)
	p.Type = paginator.Dots
	p.PerPage = 1
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(4)
	m.responsePaginator = p

	// default query empty fields
	queryInputName, queryInputValue := makeInputField("", "name"), makeInputField("", "value")
	m.queryParams.fields = append(m.queryParams.fields, queryInputName, queryInputValue)
	m.queryParams.cursor = 0
	// default header empty fields
	headerInputName, headerInputValue := makeInputField("", "header"), makeInputField("", "value")
	m.headers.fields = append(m.headers.fields, headerInputName, headerInputValue)
	m.headers.cursor = 0

	m.resHeaders = makeTable()
	m.resCookies = makeTable()

	return m
}

func makeTable() table.Model {
	columns := []table.Column{{Title: "Name", Width: 40}, {Title: "Value", Width: 40}}
	rows := []table.Row{}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false), //i dont see any difference either true or false
		table.WithHeight(7),
		table.WithWidth(80),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t

}

func setUIRequest(m *model, r *DBRequest) {
	for i, option := range m.method.options {
		if option == r.Method {
			m.method.cursor = i
			break
		}
	}
	m.url.SetValue(r.Url)

	// populate body from selected body type
	m.body.cursor = r.Body.Selected
	m.body.field.SetValue(r.Body.Types[r.Body.Selected].Value)
	for i, item := range r.Body.Types {
		m.body.options[i].value = item.Value
	}

	// populate saved auth
	for i, type_ := range r.Auth.Types {
		uiAuth := m.auth.options[i]
		for j, field := range type_.Fields {
			uiAuth.fields[j].SetValue(field.Value)
		}
	}
	m.auth.cursor = r.Auth.Selected

	// populate saved query params
    m.queryParams.fields = nil
	for _, queryParam := range r.Query {
		name := makeInputField(queryParam.Name, "name")
		value := makeInputField(queryParam.Value, "value")
		m.queryParams.fields = append(m.queryParams.fields, name, value)
	}
    m.queryParams.fields = append(m.queryParams.fields, makeInputField("", "name"), makeInputField("", "name"))

	// populate saved request headers
    m.headers.fields = nil
	for _, header := range r.Headers {
		name := makeInputField(header.Name, "header")
		value := makeInputField(header.Value, "value")
		m.headers.fields = append(m.headers.fields, name, value)
	}
    m.headers.fields = append(m.headers.fields, makeInputField("", "header"), makeInputField("", "name"))

	if len(r.Responses) > 0 {
		r := &r.Responses[0]
		m.resBody.SetContent(r.Body)
		// populate saved response headers
		res_headers := []table.Row{}
		for _, item := range r.Headers {
			res_headers = append(res_headers, table.Row{item.Name, item.Value})
		}
		m.resHeaders.SetRows(res_headers)

		// populate saved response cookies
		res_cookies := []table.Row{}
		for _, item := range r.Cookies {
			res_cookies = append(res_cookies, table.Row{item.Name, item.Value})
		}
		m.resCookies.SetRows(res_cookies)

		// populate trace logs
		m.resLogs = r.TraceLogs
	}
}
