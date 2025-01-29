package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

type gameState struct {
	screen tcell.Screen
	p1     *coord
	p2     *coord
	ball   *ball
	ticks  int
}

func (gs *gameState) tick() {
	if gs.ticks > 10 {
		gs.ticks = 0
	} else {
		gs.ticks += 1
	}
}

func (gs *gameState) moveBall() {
	gs.ball.x -= 1
}

type ball struct {
	x int
	y int
}

type coord struct {
	yTop int
	yBot int
}

func (c *coord) up() {
	c.yBot -= 1
	c.yTop -= 1
}

func (c *coord) down() {
	c.yBot += 1
	c.yTop += 1
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	gs := gameState{screen: s, p1: &coord{2, 5}, p2: &coord{2, 5}, ball: &ball{55, 20}, ticks: 0}

	s.Clear()
	s.DisableMouse()

	drawBorder(s)
	// left
	drawLeftPaddle(gs)

	// right
	drawRightPaddle(gs)

	drawBall(gs)

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

	keyEvent := make(chan tcell.Key)
	kill := make(chan bool)
	tick := make(chan int)

	go keyboardEventLoop(keyEvent, kill, gs)

	go func() {
		ticknum := 0
		for {
			time.Sleep(time.Millisecond * 100)
			tick <- ticknum
			ticknum++
		}
	}()

	// game loop
	for {
		gs.screen.Clear()

		select {
		case ev := <-keyEvent:
			if ev == tcell.KeyUp {
				gs.p1.up()
			} else {
				gs.p1.down()
			}

		case <-kill:
			return

		case <-tick:
			gs.moveBall()
		}

		drawBorder(gs.screen)
		drawBall(gs)
		drawLeftPaddle(gs)
		drawRightPaddle(gs)
		gs.screen.Show()
	}
}

func keyboardEventLoop(ch chan tcell.Key, kill chan bool, gs gameState) {
	// main event loop listening to keyboard events
	for {

		ev := gs.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			gs.screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				kill <- true
			}

			if ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown {
				ch <- ev.Key()
			}
		}
	}
}

func runGame(gs gameState) {
	for {
		time.Sleep(100 * time.Millisecond)
		gs.tick()
		gs.moveBall()
		drawBall(gs)
	}
}

func movePaddle(gs gameState, dir tcell.Key) {
	if dir == tcell.KeyUp {
		gs.p1.up()
	} else {
		gs.p1.down()
	}

	gs.screen.Show()
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

func drawLeftPaddle(gs gameState) {
	drawPaddle(gs.p1, gs.screen, 2)
}

func drawRightPaddle(gs gameState) {
	drawPaddle(gs.p2, gs.screen, 110)
}

func drawPaddle(p *coord, s tcell.Screen, x int) {
	// Left is x = 1, should start at y = 2, end at y = 6

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for i := p.yTop; i <= p.yBot; i++ {
		s.SetContent(x, i, tcell.RuneBlock, nil, style)
	}
}

func drawBall(gs gameState) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	gs.screen.SetContent(gs.ball.x, gs.ball.y, tcell.RuneDiamond, nil, style)
}
