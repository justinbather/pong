package main

import (
	"errors"
	"log"
	"os"

	"github.com/gdamore/tcell"
)

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	s.Clear()
	s.DisableMouse()

	drawBorder(s)
	// left
	drawPaddle(s, 2)
	// right
	drawPaddle(s, 110)

	drawBall(s)

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
		os.Exit(1)
	}
	defer quit()

	for {
		s.Show()
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}
}

func drawBorder(s tcell.Screen) error {
	w, h := s.Size()
	if w < 50 || h < 50 {
		return errors.New("Window too small")
	}

	maxW := 110
	maxH := 40

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// Draw top and bottom horizontal borders
	for x := 1; x <= maxW; x++ {
		s.SetContent(x, 1, tcell.RuneHLine, nil, style)      // Top border
		s.SetContent(x, maxH+1, tcell.RuneHLine, nil, style) // Bottom border
	}

	// Draw left and right vertical borders
	for y := 1; y <= maxH; y++ {
		s.SetContent(1, y, tcell.RuneVLine, nil, style)      // Left border
		s.SetContent(maxW+1, y, tcell.RuneVLine, nil, style) // Right border
	}

	// Draw corners
	s.SetContent(1, 1, tcell.RuneULCorner, nil, style)           // Upper left corner
	s.SetContent(maxW+1, 1, tcell.RuneURCorner, nil, style)      // Upper right corner
	s.SetContent(1, maxH+1, tcell.RuneLLCorner, nil, style)      // Lower left corner
	s.SetContent(maxW+1, maxH+1, tcell.RuneLRCorner, nil, style) // Lower right corner

	return nil
}

func drawPaddle(s tcell.Screen, x int) {
	// Left is x = 1, should start at y = 2, end at y = 6

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	s.SetContent(x, 2, tcell.RuneBlock, nil, style)
	s.SetContent(x, 3, tcell.RuneBlock, nil, style)
	s.SetContent(x, 4, tcell.RuneBlock, nil, style)
}

func drawBall(s tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	s.SetContent(55, 20, tcell.RuneDiamond, nil, style)
}
