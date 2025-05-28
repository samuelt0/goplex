package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("failed to initialise screen: %v", err)
	}
	defer screen.Fini()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		screen.Fini()
		os.Exit(0)
	}()

	drawSplash(screen)

	for {
		ev := screen.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
			drawSplash(screen)
		}
	}
}

func drawSplash(s tcell.Screen) {
	s.Clear()

	msg := "goplex — press Ctrl‑C to quit"
	w, h := s.Size()
	x := (w - len(msg)) / 2
	y := h / 2

	for i, r := range msg {
		s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
	s.Show()
}
