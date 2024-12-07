package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	noAuth int = iota
	basicAuth
	digestAuth
	tokenAuth
	customAuth
)

type authType struct {
	name   string
	fields []textinput.Model
	cursor int
}

type auth struct {
	options []*authType
	cursor  int
}

func makeAuth() auth {
	return auth{
		[]*authType{
			{"No Auth", []textinput.Model{}, 0},
			{"Basic Auth", []textinput.Model{makeInputField("", ""), makeInputField("", "")}, 0},
			{"Digest", []textinput.Model{makeInputField("", ""), makeInputField("", "")}, 0},
			{"Bearer Token", []textinput.Model{makeInputField("", ""), makeInputField("", "")}, 0},
			{"Custom", []textinput.Model{makeInputField("", "")}, 0},
		},
		0,
	}
}

type bodyType struct {
	type_ string
	value string
}

type body struct {
	options []bodyType
	cursor  int
	field   textarea.Model
}
 
func makeBody() body {
	t := textarea.New()
	t.BlurredStyle.CursorLineNumber = t.BlurredStyle.CursorLineNumber.Background(primaryColor)
	t.BlurredStyle.Base = t.BlurredStyle.Base.Background(primaryColor)

	t.FocusedStyle.CursorLine = t.FocusedStyle.CursorLine.Background(focusedColor)
	t.FocusedStyle.Base = t.FocusedStyle.Base.Background(focusedColor)
	t.Blur()
	t.Prompt = ""
    // if i dont set width leaves blank space in the end of each line
	t.SetWidth(rightPanelWidth - 2)
    // setting placeholder screws up the background
	// t.Placeholder = "..."
	return body{
		[]bodyType{{"No body", ""}, {"Json", ""}, {"Xml", ""}, {"Plain", ""}},
		0,
		t,
	}
}

func makeSelectField(options ...string) selectField {
	return selectField{
		options: options,
		cursor:  0,
	}
}

func makeInputField(value string, placeholder string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.SetValue(value)
	return input
}

type selectField struct {
	options []string
	cursor  int
}

type inputFields struct {
	cursor int
	fields []textinput.Model
}

func (s *selectField) down() {
	s.cursor = min(s.cursor+1, len(s.options)-1)
}

func (s *selectField) up() {
	s.cursor = max(s.cursor-1, 0)
}

func (b *body) down() {
	b.cursor = min(b.cursor+1, len(b.options)-1)
}

func (b *body) up() {
	b.cursor = max(b.cursor-1, 0)
}

func (a *auth) down() {
	a.cursor = min(a.cursor+1, len(a.options)-1)
}

func (a *auth) up() {
	a.cursor = max(a.cursor-1, 0)
}

func (a *authType) down() bool {
	if a.cursor == len(a.fields)-1 {
		return false
	}
	a.cursor++
	return true
}

func (a *authType) up() bool {
	if a.cursor == 0 {
		return false
	}
	a.cursor--
	return true
}

func (i *inputFields) down() bool {
	if i.cursor >= len(i.fields)-2 {
		return false
	}
	i.cursor += 2
	return true
}

func (i *inputFields) up() bool {
	if i.cursor < 2 {
		return false
	}
	i.cursor -= 2
	return true
}
