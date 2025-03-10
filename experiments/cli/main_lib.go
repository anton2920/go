package main

import (
	"log"
	"os"
	"os/signal"

	"golang.org/x/term"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to switch terminal to RAW mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		term.Restore(int(os.Stdin.Fd()), oldState)
		os.Exit(0)
	}()

	if err := App(); err != nil {
		log.Printf("Failed to run application: %v", err)
	}
}
