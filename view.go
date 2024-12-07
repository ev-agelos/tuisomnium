package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

var (
	rightPanelWidth  = 80
	grey             = lg.Color("237")
	white            = lg.Color("15")
	magenta          = lg.Color("57")
	green            = lg.Color("22")
	red              = lg.Color("160")
	yellow           = lg.Color("3")
	orange           = lg.Color("208")
	blue             = lg.Color("4")
	primaryColor     = lg.Color("236")
	secondaryColor   = lg.Color("234")
	focusedColor     = lg.Color("238")
	primary          = lg.NewStyle().Background(primaryColor)
	secondary        = lg.NewStyle().Background(secondaryColor)
	focused          = lg.NewStyle().Background(focusedColor).Foreground(white)
	placeHolderColor = lg.Color("241")

	nameListStyle    = primary.Width(15).Height(10).Border(lg.NormalBorder())
	nameStyle        = primary
	nameFocusedStyle = focused.Inherit(nameStyle)

	methodStyle        = lg.NewStyle().Width(9).AlignHorizontal(lg.Center).Padding(0, 1, 0)
	methodFocusedStyle = methodStyle.Inherit(focused)

	urlWidth = rightPanelWidth - methodStyle.GetWidth() - buttonStyle.GetWidth() - 3

	buttonStyle          = lg.NewStyle().Padding(0, 1, 0).Width(6)
	buttonFocusedStyle   = buttonStyle.Background(lg.Color("99")).Foreground(lg.Color("13"))
	buttonUnFocusedStyle = buttonStyle.Background(magenta).Foreground(white)

	responseStyle          = lg.NewStyle().Width(rightPanelWidth)
	responseFocusedStyle   = responseStyle.Inherit(focused)
	responseUnFocusedStyle = responseStyle.Inherit(primary)

	requestPaginatorStyle = lg.NewStyle().Width(rightPanelWidth).Height(10).Border(lg.NormalBorder())

	tableFocusedStyle   = lg.NewStyle().BorderStyle(lg.NormalBorder()).BorderForeground(focusedColor)
	tableUnFocusedStyle = lg.NewStyle().BorderStyle(lg.NormalBorder()).BorderForeground(primaryColor)

	down_arrow = lg.NewStyle().PaddingRight(1).Render("")
)

func coloredMethod(s string) string {
	var color lg.Color
	switch s {
	case "GET":
		color = magenta
	case "POST":
		color = green
	case "PUT":
		color = orange
	case "PATCH":
		color = yellow
	case "DELETE":
		color = red
	default:
		color = blue
	}
	return lg.NewStyle().Foreground(color).Render(s)
}

func getStatusStyle(s string) lg.Style {
	var color lg.Color
	switch string(s[0]) {
	case "2":
		color = green
	case "3":
		color = magenta
	case "4":
		color = orange
	case "5":
		color = red
	}
	return lg.NewStyle().Background(color)
}

func renderRequestNames(m model) string {
	pipe := primary.Foreground(magenta).Render("| ")
	var names []string
	for i, field := range m.names {
		if view == requestsView && i == m.requests.cursor {
			names = append(names, pipe+nameFocusedStyle.Render(field.View()))
		} else if i == m.requests.cursor {
			names = append(names, pipe+nameStyle.Render(field.View()))
		} else {
			names = append(names, primary.Render("  ")+nameStyle.Render(field.View()))
		}
	}
	return nameListStyle.Render(strings.Join(names, "\n"))
}

func renderMethod(f selectField) string {
	switch view {
	case httpMethodView:
		return methodStyle.Inherit(focused).Render(coloredMethod(f.options[f.cursor])) + focused.Render(down_arrow)
	default:
		return methodStyle.Inherit(secondary).Render(coloredMethod(f.options[f.cursor])) + secondary.Render(down_arrow)
	}
}

func renderMethodOptions(m model) string {
	if view != methodOptionsView {
		return ""
	}
	options := []string{}
	for i, option := range m.method.options {
		cursor := "  "
		if m.method.cursor == i {
			cursor = "> "
		}
		options = append(options, cursor+coloredMethod(option))
	}
	return secondary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(strings.Join(options, "\n"))
}

func renderUrl(f textinput.Model) string {
	f.Width = urlWidth
	switch view {
	case urlView:
		f.Cursor.TextStyle = f.Cursor.TextStyle.Background(focusedColor)
		f.TextStyle = f.TextStyle.Foreground(white).Background(focusedColor)
		f.PlaceholderStyle = f.PlaceholderStyle.Background(focusedColor)
	default:
		f.Cursor.TextStyle = f.Cursor.TextStyle.Background(secondaryColor)
		f.TextStyle = f.TextStyle.UnsetForeground().Background(secondaryColor)
		f.PlaceholderStyle = f.PlaceholderStyle.Background(secondaryColor)
	}
	return f.View()
}

func renderSendButton() string {
	switch view {
	case sendView:
		return buttonFocusedStyle.Render("Send")
	default:
		return buttonUnFocusedStyle.Render("Send")
	}
}

func renderBodySetting(m model) string {
	style := lg.NewStyle().Padding(0, 1, 0).AlignHorizontal(lg.Center).Width(9)
	switch view {
	case bodyTypeView:
		return style.Inherit(focused).Render(m.body.options[m.body.cursor].type_) + focused.Render(down_arrow)
	default:
		return style.Inherit(primary).Render(m.body.options[m.body.cursor].type_) + primary.Render(down_arrow)
	}
}

func renderBodyOptions(m model) string {
	if view != bodyOptionsView {
		return ""
	}
	list_style := primary.Padding(0, 1, 0)
	options := []string{}
	for i, option := range m.body.options {
		cursor := "  "
		if m.body.cursor == i {
			cursor = "> "
		}
		options = append(options, list_style.Render(cursor+option.type_))
	}
	return strings.Join(options, "\n")

}

func renderAuthSetting(m model) string {
	style := lg.NewStyle().Padding(0, 1, 0).AlignHorizontal(lg.Center).Width(14)
	switch view {
	case authTypeView:
		return style.Inherit(focused).Render(m.auth.options[m.auth.cursor].name) + focused.Render(down_arrow)
	default:
		return style.Inherit(primary).Render(m.auth.options[m.auth.cursor].name) + primary.Render(down_arrow)
	}
}

func renderAuthOptions(m model) string {
	if view != authOptionsView {
		return ""
	}
	style := lg.NewStyle().Padding(0, 1, 0).Width(14)
	list_style := primary.Padding(0, 1, 0).Width(style.GetWidth() + 2) // +2 because doesnt include down_arrow
	options := []string{}
	for i, option := range m.auth.options {
		cursor := "  "
		if m.auth.cursor == i {
			cursor = "> "
		}
		options = append(options, list_style.Render(cursor+option.name))
	}
	return strings.Join(options, "\n")

}
func renderSetting(s string, focusedView int) string {
	switch view {
	case focusedView:
		return focused.Padding(0, 1, 0).Render(s)
	default:
		return primary.Padding(0, 1, 0).Render(s)
	}
}

func renderRequestBody(m model) string {
	if m.body.cursor == 0 {
		width := requestPaginatorStyle.GetWidth()
		height := requestPaginatorStyle.GetHeight() - 5 // FIXME why -5? need to find out
		return lg.NewStyle().Width(width).Height(height).Align(lg.Center, lg.Center).Render("Select body from above")
	}

	switch view {
	case bodyContentView:
        if !m.body.field.Focused(){
            m.body.field.Blur()  // if not blur() colors dont change no matter what
            m.body.field.BlurredStyle.CursorLineNumber = m.body.field.BlurredStyle.CursorLineNumber.Background(focusedColor)
            m.body.field.BlurredStyle.Base = m.body.field.BlurredStyle.Base.Background(focusedColor)
        }
	}
	return m.body.field.View()
}

func getAuthView(m model) string {
	style := primary.Border(lg.NormalBorder(), false, false, true, false).Width(20)
	var prefix0, prefix1 string
	auth := m.auth.options[m.auth.cursor]
	switch m.auth.cursor {
	case noAuth:
		width := requestPaginatorStyle.GetWidth()
		height := requestPaginatorStyle.GetHeight()
		return lg.NewStyle().Width(width).Height(height).Align(lg.Center, lg.Center).Render("󰌿\n\nSelect an auth type from above")
	case customAuth:
		if view == authContentView && auth.cursor == 0 {
			style = style.Inherit(focused)
		}
		return lg.JoinHorizontal(lg.Left, "Custom ", style.Width(50).Render(auth.fields[0].View()))
	case basicAuth, digestAuth:
		prefix0, prefix1 = "Username ", "Password "
	case tokenAuth:
		prefix0, prefix1 = "Prefix ", "Token  "
	}

	s0, s1 := style.BorderBottomForeground(placeHolderColor), style.BorderBottomForeground(placeHolderColor)
	if view == authContentView {
		if auth.cursor == 0 {
			s0 = s0.BorderBottomForeground(white)//Inherit(focused)
		}
		if auth.cursor == 1 {
			s1 = s1.BorderBottomForeground(white)//Inherit(focused)
		}
	}
	return lg.JoinHorizontal(lg.Left, prefix0, s0.Render(auth.fields[0].View())) + "\n" + lg.JoinHorizontal(lg.Left, prefix1, s1.Render(auth.fields[1].View()))
}

func renderNameValueFields(f inputFields, focusedView int) string {
	var lines []string
	var rendered string
	width := requestPaginatorStyle.GetWidth() / 2
	for i, field := range f.fields {
        field.PlaceholderStyle = field.PlaceholderStyle.Background(primaryColor)
		style := primary.Width(width-20).Border(lg.NormalBorder(), false, false, true, false)
		if view == focusedView && f.cursor == i {
			rendered = style.BorderBottomForeground(white).BorderBottomBackground(primaryColor).Render(field.View())
		} else {
			rendered = style.BorderBottomForeground(placeHolderColor).BorderBottomBackground(primaryColor).Render(field.View())
		}
		// FIXME dont join with space, measure width to show it like the table
		if i%2 != 0 {
			lastIndex := len(lines) - 1
			// have to add border to spaces because JoinHorizontal brakes colors
			spaces := primary.Foreground(primaryColor).Border(lg.NormalBorder(), false, false, true, false).BorderBottomForeground(primaryColor).BorderBottomBackground(primaryColor).Render("   ")
			lines[lastIndex] = lg.JoinHorizontal(lg.Center, lines[lastIndex], spaces, rendered)
		} else {
			lines = append(lines, rendered)
		}
	}
	return strings.Join(lines, "\n")
}

func renderRelativeTime(m model) string {
	if len(m.requests.items[m.requests.cursor].Responses) == 0 {
		return ""
	}
	res := &m.requests.items[m.requests.cursor].Responses[m.history.cursor]
	relativeTime := ""
	if !res.ResponseAt.IsZero() {
		relativeTime = lg.NewStyle().Padding(0, 1, 0).Render(humanize.Time(res.ResponseAt) + " ")
	}
	if view == historyView {
		return focused.Render(relativeTime)
	}
	return secondary.Render(relativeTime)
}

func renderLeftStatusLine(m model) (string, string, string) {
	status, duration, size := "", "", ""
	if m.requests.items[m.requests.cursor] != nil && len(m.requests.items[m.requests.cursor].Responses) > 0 {
		res := &m.requests.items[m.requests.cursor].Responses[m.history.cursor]
		status = getStatusStyle(res.Status).Padding(0, 1, 0).Render(res.Status)
		duration = primary.Padding(0, 1, 0).Render(res.Duration)
		size = primary.Padding(0, 1, 0).Render(res.Size + " B")
	}
	return status, duration, size
}

func renderStatusLine(m model) string {
	status, duration, size := renderLeftStatusLine(m)
	relativeTime := renderRelativeTime(m)
	combined := status + secondary.Render(" ") + duration + secondary.Render(" ") + size
	padding := rightPanelWidth - lg.Width(combined) - lg.Width(relativeTime)
	padding = max(padding, 0) // Avoid negative padding
	combined = secondary.PaddingLeft(1).PaddingRight(padding - 1).Render(combined)
	return combined + relativeTime
}

func renderHistory(m model) string {
	if len(m.requests.items[m.requests.cursor].Responses) == 0 || view != historyOptionsView {
		return ""
	}
	space := primary.Render(" ")
	var b strings.Builder
	b.WriteString("\n")
	for i, response := range m.requests.items[m.requests.cursor].Responses {
		cursor := "  "
		if m.history.cursor == i {
			cursor = "> "
		}
		status := getStatusStyle(response.Status).Render(response.Status)
		method_with_url := focused.Render(response.RequestMethod + " " + response.RequestUrl)
		duration := focused.Render(response.Duration)
		size := focused.Render(response.Size)
		line := cursor + strings.Join([]string{status, method_with_url, duration, size}, space)
		b.WriteString("\n" + line)
	}
	return b.String()
}

func renderResponseBody(m model) string {
	switch view {
	case responseView:
		return responseFocusedStyle.Render(m.resBody.View())
	default:
		return responseUnFocusedStyle.Render(m.resBody.View())
	}
}

func renderResponseHeaders(m model) string {
	switch view {
	case responseHeadersView:
		return tableFocusedStyle.Render(m.resHeaders.View())
	default:
		return tableUnFocusedStyle.Render(m.resHeaders.View())
	}
}

func renderResponseCookies(m model) string {
	switch view {
	case responseCookiesView:
		return tableFocusedStyle.Render(m.resCookies.View())
	default:
		return tableUnFocusedStyle.Render(m.resCookies.View())
	}
}

func (m model) View() string {
	helpKeys := " (n) 󰆴 (d)\n"
	requestNames := renderRequestNames(m)
	if len(m.requests.items) == 0 {
		leftSide := helpKeys + "\n" + requestNames
		rightTop := secondary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render("")
		requestSettings := primary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render("")
		rightSide := rightTop + "\n" + requestSettings
		request_paginator := requestPaginatorStyle.Inherit(primary).Render("")
		status_line := secondary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render("")
		rightSide += "\n" + request_paginator + "\n" + status_line + "\n" + ""

		return lg.NewStyle().Render(lg.JoinHorizontal(lg.Top, leftSide, rightSide))
	}
	method := renderMethod(m.method)
	methodOptions := renderMethodOptions(m)
	url := renderUrl(m.url)
	button := renderSendButton()
	body_setting := renderBodySetting(m)
	body_options := renderBodyOptions(m)
	auth_setting := renderAuthSetting(m)
	auth_options := renderAuthOptions(m)
	query_setting := renderSetting("Query", queryView)
	header_setting := renderSetting("Headers", headersView)

	status_line := secondary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(renderStatusLine(m) + renderHistory(m))

	leftSide := helpKeys + "\n" + requestNames
	rightTop := secondary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(lg.JoinHorizontal(lg.Top, method, url, button))
	requestSettings := primary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(body_setting + auth_setting + query_setting + header_setting)

	var content string
	switch m.requestPaginator.Page {
	case 0:
		content = renderRequestBody(m)
	case 1:
		content = getAuthView(m)
	case 2:
		content = renderNameValueFields(m.queryParams, queryContentView)
	case 3:
		content = renderNameValueFields(m.headers, headersContentView)
	}
	request_paginator := requestPaginatorStyle.Inherit(primary).Render("\n" + content + "\n\n" + m.requestPaginator.View())

	switch m.responsePaginator.Page {
	case 0:
		content = renderResponseBody(m)
	case 1:
		content = renderResponseHeaders(m)
	case 2:
		content = renderResponseCookies(m)
	case 3:
		content = lg.NewStyle().Border(lg.NormalBorder()).Render(m.resLogs)
	}
	response_paginator := content + "\n\n" + m.responsePaginator.View()

	rightSide := rightTop
	if len(methodOptions) > 0 {
		rightSide += "\n" + primary.Render(methodOptions)
	}
	rightSide += "\n" + requestSettings

	if len(body_options) > 0 {
		rightSide += "\n" + primary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(body_options)
	} else if len(auth_options) > 0 {
		rightSide += "\n" + primary.Width(rightPanelWidth).Border(lg.NormalBorder()).Render(auth_options)
	}
	rightSide += "\n" + request_paginator + "\n" + status_line + "\n" + response_paginator

	return lg.NewStyle().Render(lg.JoinHorizontal(lg.Top, leftSide, rightSide))
}
