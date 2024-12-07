package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	store := new(Store)
	if err := store.Init(); err != nil {
		log.Fatalf("unable to init store: %v", err)
	}
	p := tea.NewProgram(initialModel(store))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting app: %v", err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
