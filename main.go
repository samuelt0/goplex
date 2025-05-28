package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell"

	"goplex/pane"
)

var curX, curY int

func main() {
	scr, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("new screen: %v", err)
	}
	if err := scr.Init(); err != nil {
		log.Fatalf("screen init: %v", err)
	}
	defer scr.Fini()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		scr.Fini()
		os.Exit(0)
	}()

	p, err := pane.NewPane(os.Getenv("SHELL"))
	if err != nil {
		log.Fatalf("pane: %v", err)
	}
	defer p.Close()

	go func() {
		for data := range p.Out {
			drawBytes(scr, data)
		}
	}()

	scr.Clear()
	scr.Show()

	for {
		ev := scr.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventResize:
			scr.Sync()
		case *tcell.EventKey:
			if e.Key() == tcell.KeyCtrlC {
				return
			}
			if r := e.Rune(); r != 0 {
				_ = p.WriteRune(r)
			} else {
				switch e.Key() {
				case tcell.KeyEnter:
					_ = p.WriteRune('\n')
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					_ = p.WriteRune('\b')
				case tcell.KeyTab:
					_ = p.WriteRune('\t')
				}
			}
		}
	}
}

func drawBytes(scr tcell.Screen, data []byte) {
	w, h := scr.Size()

	for _, b := range data {
		if b == '\n' {
			curX = 0
			curY++
			if curY >= h {
				scrollUp(scr, w, h)
				curY = h - 1
			}
			continue
		}

		scr.SetContent(curX, curY, rune(b), nil, tcell.StyleDefault)
		curX++
		if curX >= w {
			curX = 0
			curY++
			if curY >= h {
				scrollUp(scr, w, h)
				curY = h - 1
			}
		}
	}
	scr.Show()
}

func scrollUp(scr tcell.Screen, w, h int) {
	for row := 1; row < h; row++ {
		for col := 0; col < w; col++ {
			mainc, comb, style, _ := scr.GetContent(col, row)
			scr.SetContent(col, row-1, mainc, comb, style)
		}
	}
	for col := 0; col < w; col++ {
		scr.SetContent(col, h-1, ' ', nil, tcell.StyleDefault)
	}
}
