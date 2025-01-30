package main

import (
	"errors"
	"log"
	"math"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

const TICK_RATE = 100

type gameState struct {
	screen tcell.Screen
	p1     *coord
	p2     *coord
	ball   *ball
	ticks  int
}

func (gs *gameState) moveBall() {
	dir := gs.ball.dir

	if dir == "L" || dir == "TL" || dir == "BL" {
		collision, newDir := calculateCollision(*gs.ball, *gs.p1)
		if collision {
			gs.ball.dir = newDir
		} else {
			gs.ball.dir = dir
		}

		x, y := calcCoord(*gs.ball)
		gs.ball.x = x
		gs.ball.y = y
	}

	// collision
	if dir == "R" || dir == "TR" || dir == "BR" {
		collision, newDir := calculateCollision(*gs.ball, *gs.p2)
		if collision {
			gs.ball.dir = newDir
		} else {
			gs.ball.dir = dir
		}

		x, y := calcCoord(*gs.ball)
		gs.ball.x = x
		gs.ball.y = y
	}

	log.Printf("Set direction from %s to %s at x: %d y: %d", dir, gs.ball.dir, gs.ball.x, gs.ball.y)
}

func calcCoord(ball ball) (int, int) {
	switch ball.dir {
	case "L":
		return ball.x - 1, ball.y
	case "R":
		return ball.x, ball.y + 1
	case "TL":
		return ball.y - 1, ball.y - 1
	case "TR":
		return ball.x + 1, ball.y - 1
	case "BL":
		return ball.x - 1, ball.y + 1
	case "BR":
		return ball.x + 1, ball.y + 1
	default:
		return 2, 2
	}
}

type ball struct {
	x   int
	y   int
	dir string
}

type coord struct {
	yTop int
	yBot int
	x    int
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
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set output of logs to file
	log.SetOutput(logFile)
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	gs := gameState{screen: s, p1: &coord{2, 5, 2}, p2: &coord{2, 5, 110}, ball: &ball{55, 20, "L"}, ticks: 0}

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
			time.Sleep(time.Millisecond * TICK_RATE)
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

// returns string direction
// "L" left, "R" - right, "TL" top-left, "TR" top-right, "BL" bottom-left, "RL" right-left
func calculateCollision(ball ball, paddle coord) (bool, string) {
	// collided with top or bottom, reflect
	if ball.y == 2 || ball.y == 40 {
		return true, getOpposite(ball.dir)
	}

	// collision on paddle
	//left paddle
	log.Printf("paddle: x: %d, top: %d, bott: %d", paddle.x, paddle.yTop, paddle.yBot)
	if (ball.y <= paddle.yBot && ball.y >= paddle.yTop) && (ball.x == paddle.x+1 || ball.x == paddle.x-1) {
		log.Printf("Calcing collision\npaddle: %v\n ball: %v\n", paddle, ball)
		if ball.x == paddle.x+1 {
			ball.dir = calcPaddleCollision(paddle, ball, "L")
			return true, ball.dir
		} else if ball.x == paddle.x-1 {
			ball.dir = calcPaddleCollision(paddle, ball, "R")
			return true, ball.dir
		}

	}
	return false, ""
}

func calcPaddleCollision(paddle coord, ball ball, side string) string {
	log.Printf("calculating paddle coll. \n\tpaddle: %+v\n\tball: %+v\n\tside: %s\n", paddle, ball, side)
	mid := math.Round(float64(paddle.yBot-paddle.yTop) / 2)
	if ball.y == int(mid) {
		if side == "L" {
			return "R"
		}
		if side == "R" {
			return "L"
		}
	}

	if ball.y > int(mid) {
		opp := oppositeAngle(ball.dir)
		if opp == "" {
			opp = oppositeNonAngle(ball.dir, "T")
		}
		return opp
	} else {
		if ball.y < int(mid) {
			opp := oppositeAngle(ball.dir)
			if opp == "" {
				opp = oppositeNonAngle(ball.dir, "B")
			}
			return opp
		}
	}

	return ""
}

func oppositeNonAngle(dir string, y string) string {
	switch dir {
	case "L":
		if y == "T" {
			return "TR"
		} else {
			return "BR"
		}
	default:
		if y == "T" {
			return "TL"
		} else {
			return "BL"
		}
	}
}

func oppositeAngle(dir string) string {
	switch dir {

	case "TL":
		return "TR"
	case "TR":
		return "TL"
	case "BR":
		return "BL"
	case "BL":
		return "BR"

	default:
		return ""

	}
}

func getOpposite(dir string) string {
	switch dir {
	case "L":
		return "R"
	case "R":
		return "L"
	case "TL":
		return "BL"
	case "TR":
		return "BR"
	case "BR":
		return "TR"
	case "BL":
		return "TL"

	default:
		return ""
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
	drawPaddle(gs.p1, gs.screen)
}

// "AI" matches the balls y value for now
func drawRightPaddle(gs gameState) {
	drawPaddle(&coord{gs.ball.y - 2, gs.ball.y + 1, 110}, gs.screen)
	gs.p2.yTop = gs.ball.y - 2
	gs.p2.yBot = gs.ball.y + 1
}

func drawPaddle(p *coord, s tcell.Screen) {
	// Left is x = 1, should start at y = 2, end at y = 6

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for i := p.yTop; i <= p.yBot; i++ {
		s.SetContent(p.x, i, tcell.RuneBlock, nil, style)
	}
}

func drawBall(gs gameState) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	gs.screen.SetContent(gs.ball.x, gs.ball.y, tcell.RuneDiamond, nil, style)
}
