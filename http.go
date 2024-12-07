package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/icholy/digest"
)

type (
	response struct {
		body          string
		status        string
		headers       []NameValue
		cookies       []NameValue
		contentLength string
		duration      string
		responseAt    time.Time
		traceLogs     string
	}
	errMsg struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

func sendRequest(m model) tea.Cmd {
	return func() tea.Msg {
		c := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest(
			m.method.options[m.method.cursor],
			m.url.Value(),
			strings.NewReader(m.body.field.Value()),
		)

		traceLogs := ""
		trace := getTrace(&traceLogs)
		ctx := httptrace.WithClientTrace(req.Context(), trace)
		req = req.WithContext(ctx)

		// auth
		auth := m.auth.options[m.auth.cursor]
		switch m.auth.cursor {
		case noAuth:
		case basicAuth:
			username, password := auth.fields[0].Value(), auth.fields[1].Value()
			req.SetBasicAuth(username, password)
		case digestAuth:
			username, password := auth.fields[0].Value(), auth.fields[1].Value()
			c.Transport = &digest.Transport{Username: username, Password: password}
		case tokenAuth:
			req.Header.Add("Authorization", auth.fields[0].Value()+" "+auth.fields[1].Value())
		case customAuth:
			req.Header.Add("Authorization", auth.fields[0].Value())
		}

		// headers
		for i := range len(m.headers.fields) / 2 {
			name := m.headers.fields[i*2].Value()
			value := m.headers.fields[i*2+1].Value()
			if len(name) > 0 && len(value) > 0 {
				req.Header.Add(name, value)
			}
		}

		// query params
		q := req.URL.Query()
		for i := range len(m.queryParams.fields) / 2 {
			name := m.queryParams.fields[i*2].Value()
			value := m.queryParams.fields[i*2+1].Value()
			if len(name) > 0 {
				q.Add(name, value)
			}
		}

		start := time.Now()
		res, err := c.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return errMsg{err}
		}
		stop := time.Now()
		duration := stop.Sub(start)

		length := "Unknown"
		if res.ContentLength != -1 {
			length = strconv.FormatInt(res.ContentLength, 10)
		}
		var headers []NameValue
		for name, values := range res.Header {
			headers = append(headers, NameValue{name, strings.Join(values, ",")})
		}
		var cookies []NameValue
		for _, cookie := range res.Cookies() {
			cookies = append(cookies, NameValue{cookie.Name, cookie.Value})
		}

		return response{
			string(body),
			res.Status,
			headers,
			cookies,
			length,
			duration.Round(time.Millisecond).String(),
			stop,
			traceLogs,
		}
	}
}

func getTrace(b *string) *httptrace.ClientTrace {
	var start time.Time

	t0 := time.Now()
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			start = time.Now()
			*b += fmt.Sprintf("GetConn(%s) %d ms\n", hostPort, start.Sub(t0).Milliseconds())
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			start = time.Now()
			*b += fmt.Sprintf("DNSStart(%+v) %d ms\n", info, start.Sub(t0).Milliseconds())
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			start = time.Now()
			*b += fmt.Sprintf("DNSDone(%+v) %d ms\n", info, start.Sub(t0).Milliseconds())
		},
		ConnectStart: func(network, addr string) {
			*b += fmt.Sprintf("ConnectStart(%s, %s)\n", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			*b += fmt.Sprintf("ConnectDone(%s, %s, %v)\n", network, addr, err)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			start = time.Now()
			*b += fmt.Sprintf("GotConn(%+v) %d ms\n", info, start.Sub(t0).Milliseconds())
		},
		GotFirstResponseByte: func() {
			start = time.Now()
			*b += fmt.Sprintf("GotFirstResponseByte %d ms\n", start.Sub(t0).Milliseconds())
		},
		PutIdleConn: func(err error) {
			*b += fmt.Sprintf("PutIdleConn(%+v)\n", err)
		},
	}
}
