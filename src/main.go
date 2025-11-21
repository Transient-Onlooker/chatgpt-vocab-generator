package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// If you want to log to a file, you can do something like this.
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("could not create log file:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("Starting...")

	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
