package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

// how many millisecs between ticks
const TICK_RATE = 75

// dimensions
const MAX_HEIGHT = 40
const MAX_WIDTH = 110

// haha this is basically c now
// when the game starts this loads based on screen size
var GAME_TOP = 0
var GAME_BOTTOM = 0
var GAME_LEFT = 0
var GAME_RIGHT = 0
var GAME_MID_X = 0
var GAME_MID_Y = 0

const PADDLE_WIDTH = 3

// how many ticks for countdown lock to be unlocked
// makeshift CountdownLatch - not thread safe at all but thats ok
const LOCK = 10
const UNLOCKED = 0

type dir int

const (
	_ dir = iota
	UP
	DOWN
	LEFT
	RIGHT
	UP_RIGHT
	DOWN_RIGHT
	UP_LEFT
	DOWN_LEFT
)

func (d dir) String() string {
	switch d {
	case UP:
		return "UP"
	case DOWN:
		return "DOWN"
	case LEFT:
		return "LEFT"
	case RIGHT:
		return "RIGHT"
	case UP_RIGHT:
		return "UP_RIGHT"
	case DOWN_RIGHT:
		return "DOWN_RIGHT"
	case UP_LEFT:
		return "UP_LEFT"
	case DOWN_LEFT:
		return "DOWN_LEFT"
	default:
		return "UNKNOWN"
	}
}

func Dir(s int) dir {
	switch s {
	case 1:
		return UP
	case 2:
		return DOWN
	case 3:
		return LEFT
	case 4:
		return RIGHT
	case 5:
		return UP_RIGHT
	case 6:
		return DOWN_RIGHT
	case 7:
		return UP_LEFT
	default:
		return DOWN_LEFT
	}
}

type gameState struct {
	screen tcell.Screen
	p1     *player
	p2     *player
	ball   *ball
	ticks  int
}

func getRight(s tcell.Screen) int {
	x, _ := s.Size()
	return x/2 + MAX_WIDTH/2
}

func getLeft(s tcell.Screen) int {
	x, _ := s.Size()
	return x/2 - MAX_WIDTH/2
}

func getTop(s tcell.Screen) int {
	_, y := s.Size()
	return y/2 - MAX_HEIGHT/2
}

func getBottom(s tcell.Screen) int {
	_, y := s.Size()
	return y/2 + MAX_HEIGHT/2
}

func newGame(s tcell.Screen) gameState {
	return gameState{screen: s, p1: newPlayer(s, getLeft(s)+1), p2: newPlayer(s, getRight(s)-1), ball: &ball{GAME_MID_X, GAME_MID_Y, randomDirection(), UNLOCKED}, ticks: 0}
}

type player struct {
	score  int
	paddle *paddle
}

func newPlayer(s tcell.Screen, x int) *player {
	_, y := s.Size()
	return &player{score: 0, paddle: &paddle{y / 2, y/2 + PADDLE_WIDTH, x}}
}

func (gs *gameState) moveBall() {
	var paddle *paddle
	origDir := gs.ball.dir

	switch getGeneralDirection(gs.ball.dir) {
	case LEFT:
		paddle = gs.p1.paddle
	default:
		paddle = gs.p2.paddle
	}

	coll, newDir := calculateCollision(*gs.ball, *paddle)
	if coll {
		gs.ball.dir = newDir
	}

	x, y := gs.ball.next()
	gs.ball.x = x
	gs.ball.y = y

	if x <= GAME_LEFT || x >= GAME_RIGHT {
		if x <= GAME_LEFT {
			gs.p2.score += 1
		} else {
			gs.p1.score += 1
		}
		gs.reset()
	}

	log.Printf("Set direction from %s to %s at x: %d y: %d", origDir, gs.ball.dir, gs.ball.x, gs.ball.y)
}

func (gs *gameState) reset() {
	gs.ball = &ball{GAME_MID_X, GAME_MID_Y, randomDirection(), LOCK}

}

func getGeneralDirection(dir dir) dir {
	switch dir {
	case LEFT, UP_LEFT, DOWN_LEFT:
		return LEFT
	default:
		return RIGHT
	}
}

func (b *ball) next() (int, int) {
	switch b.dir {
	case LEFT:
		return b.x - 1, b.y
	case RIGHT:
		return b.x + 1, b.y
	case UP_LEFT:
		return b.x - 1, b.y - 1
	case UP_RIGHT:
		return b.x + 1, b.y - 1
	case DOWN_LEFT:
		return b.x - 1, b.y + 1
	case DOWN_RIGHT:
		return b.x + 1, b.y + 1
	default:
		log.Printf("Direction was %s when getting next ball coord", b.dir)
		return 2, 2
	}
}

type ball struct {
	x    int
	y    int
	dir  dir
	lock int
}

func (b *ball) locked() bool {
	if b.lock > 0 {
		return true
	}

	return false
}

func (b *ball) countdown() {
	if b.lock > 0 {
		b.lock -= 1
	}
}

type paddle struct {
	yTop int
	yBot int
	x    int
}

func (c *paddle) up() {
	c.yBot -= 1
	c.yTop -= 1
}

func (c *paddle) down() {
	c.yBot += 1
	c.yTop += 1
}

func initGame(gs gameState) {
	gs.screen.Clear()
	gs.screen.DisableMouse()

	drawBorder(gs.screen)
	drawLeftPaddle(gs)
	drawRightPaddle(gs)
	drawBall(gs)
}

func initArena(s tcell.Screen) {
	top, bot, left, right := getBoardDimensions(s)
	GAME_TOP = top
	GAME_BOTTOM = bot
	GAME_LEFT = left
	GAME_RIGHT = right

	x, y := s.Size()
	GAME_MID_X = x / 2
	GAME_MID_Y = y / 2

}

// returns top, bottom, left, right
func getBoardDimensions(s tcell.Screen) (int, int, int, int) {
	top := getTop(s)
	bot := getBottom(s)
	left := getLeft(s)
	right := getRight(s)

	return top, bot, left, right
}

func randomDirection() dir {
	randy := rand.Intn(8) + 1
	if randy == 0 {
		randy = 1
	}
	return Dir(randy)
}

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	initArena(s)

	gs := newGame(s)

	initGame(gs)

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
	go startTicker(tick)

	for {
		gs.screen.Clear()

		select {
		case <-kill:
			return
		case ev := <-keyEvent:
			if ev == tcell.KeyUp {
				if gs.p1.paddle.yTop > GAME_TOP+1 {
					gs.p1.paddle.up()
				}
			} else {
				if gs.p1.paddle.yBot < GAME_BOTTOM-1 {
					gs.p1.paddle.down()
				}
			}
		case <-tick:
			if !gs.ball.locked() {
				gs.moveBall()
			} else {
				gs.ball.countdown()
			}
		}

		drawScore(gs.screen, gs.p1.score, gs.p2.score)
		drawBorder(gs.screen)
		drawBall(gs)
		drawLeftPaddle(gs)
		drawRightPaddle(gs)
		gs.screen.Show()
	}
}

func startTicker(tick chan int) {
	ticknum := 0
	for {
		time.Sleep(time.Millisecond * TICK_RATE)
		tick <- ticknum
		ticknum++
	}
}

func keyboardEventLoop(ch chan tcell.Key, kill chan bool, gs gameState) {
	// main event loop listening to keyboard events
	for {
		ev := gs.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			gs.screen.Sync()
			initArena(gs.screen)
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
func calculateCollision(ball ball, paddle paddle) (bool, dir) {
	// collided with top or bottom (offset by 1 so it doesnt go into the border), reflect
	if ball.y == GAME_TOP+1 || ball.y == GAME_BOTTOM-1 {
		return true, getOpposite(ball.dir)
	}

	// collision on paddle
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
	return false, 0
}

func calcPaddleCollision(paddle paddle, ball ball, side string) dir {
	log.Printf("calculating paddle coll. \n\tpaddle: %+v\n\tball: %+v\n\tside: %s\n", paddle, ball, side)
	mid := math.Round(float64(paddle.yBot-paddle.yTop) / 2)
	if ball.y == int(mid) {
		if side == "L" {
			return RIGHT
		}
		if side == "R" {
			return LEFT
		}
	}

	if ball.y > int(mid) {
		opp := oppositeAngle(ball.dir)
		if opp == 0 {
			opp = oppositeNonAngle(ball.dir, "T")
		}

		return opp
	} else {
		opp := oppositeAngle(ball.dir)
		if opp == 0 {
			opp = oppositeNonAngle(ball.dir, "B")
		}

		return opp
	}
}

func oppositeNonAngle(dir dir, y string) dir {
	switch dir {
	case LEFT:
		if y == "T" {
			return UP_RIGHT
		} else {
			return DOWN_RIGHT
		}
	default:
		if y == "T" {
			return UP_LEFT
		} else {
			return DOWN_LEFT
		}
	}
}

func oppositeAngle(dir dir) dir {
	switch dir {
	case UP_LEFT:
		return UP_RIGHT
	case UP_RIGHT:
		return UP_LEFT
	case DOWN_RIGHT:
		return DOWN_LEFT
	case DOWN_LEFT:
		return DOWN_RIGHT
	default:
		return 0
	}
}

func getOpposite(dir dir) dir {
	switch dir {
	case LEFT:
		return RIGHT
	case RIGHT:
		return LEFT
	case UP_LEFT:
		return DOWN_LEFT
	case UP_RIGHT:
		return DOWN_RIGHT
	case DOWN_RIGHT:
		return UP_RIGHT
	default:
		return UP_LEFT
	}
}

func movePaddle(gs gameState, dir tcell.Key) {
	if dir == tcell.KeyUp {
		gs.p1.paddle.up()
	} else {
		gs.p1.paddle.down()
	}

	gs.screen.Show()
}

func drawScore(s tcell.Screen, p1, p2 int) {
	// under the border for now?
	y := GAME_BOTTOM + 2
	//mid := 55
	// score

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	s.SetContent(GAME_MID_X-2, y, 'S', nil, style)
	s.SetContent(GAME_MID_X-1, y, 'C', nil, style)
	s.SetContent(GAME_MID_X, y, 'O', nil, style)
	s.SetContent(GAME_MID_X+1, y, 'R', nil, style)
	s.SetContent(GAME_MID_X+2, y, 'E', nil, style)

	// p1
	p1Score := getRuneScore(p1)
	for i, r := range p1Score {
		s.SetContent(GAME_MID_X-8+i, y+1, r, nil, style)
	}

	p2Score := getRuneScore(p2)
	for i, r := range p2Score {
		s.SetContent(GAME_MID_X+6+i, y+1, r, nil, style)
	}
}

func getRuneScore(score int) []rune {
	scoreStr := fmt.Sprintf("%03d", score) // Format to always be 3 digits
	return []rune(scoreStr)
}

func drawBorder(s tcell.Screen) error {
	w, h := s.Size()
	if w < 150 || h < 60 {
		return errors.New("Window too small")
	}

	midY := h / 2
	midX := w / 2

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	/*
		* To calculate the placement of the borders we get the middle x and y
		 for x axis, we use the middle, minus half of our board size, and draw over until mid x + half of board size

		* we then bump the borders out by 1 to accomodate for the corners

		* repeat on y axis
	*/

	// Draw top and bottom horizontal borders
	for x := midX - MAX_WIDTH/2; x <= MAX_WIDTH/2+midX; x++ {
		s.SetContent(x, GAME_TOP, tcell.RuneHLine, nil, style)    // Top border
		s.SetContent(x, GAME_BOTTOM, tcell.RuneHLine, nil, style) // Bottom border
	}

	// Draw left and right vertical borders
	for y := midY - MAX_HEIGHT/2; y <= MAX_HEIGHT/2+midY; y++ {
		s.SetContent(GAME_LEFT, y, tcell.RuneVLine, nil, style)  // Left border
		s.SetContent(GAME_RIGHT, y, tcell.RuneVLine, nil, style) // Right border
	}

	// Draw corners
	s.SetContent(GAME_LEFT, GAME_TOP, tcell.RuneULCorner, nil, style)     // Upper left corner
	s.SetContent(GAME_RIGHT, GAME_TOP, tcell.RuneURCorner, nil, style)    // Upper right corner
	s.SetContent(GAME_LEFT, GAME_BOTTOM, tcell.RuneLLCorner, nil, style)  // Lower left corner
	s.SetContent(GAME_RIGHT, GAME_BOTTOM, tcell.RuneLRCorner, nil, style) // Lower right corner

	return nil
}

func drawLeftPaddle(gs gameState) {
	drawPaddle(gs.p1.paddle, gs.screen)
}

// "AI" matches the balls y value for now
func drawRightPaddle(gs gameState) {
	yTop := gs.ball.y - 2
	yBot := gs.ball.y + 1

	// check bounds
	if yTop <= GAME_TOP+1 {
		drawPaddle(&paddle{GAME_TOP + 1, GAME_TOP + 4, gs.p2.paddle.x}, gs.screen)
		gs.p2.paddle.yTop = GAME_TOP + 1
		gs.p2.paddle.yBot = GAME_TOP + 4
		return
	} else if yBot >= GAME_BOTTOM-1 {
		drawPaddle(&paddle{GAME_BOTTOM - 4, GAME_BOTTOM - 1, gs.p2.paddle.x}, gs.screen)
		gs.p2.paddle.yTop = GAME_BOTTOM - 4
		gs.p2.paddle.yBot = GAME_BOTTOM - 1
		return
	} else {
		drawPaddle(&paddle{gs.ball.y - 2, gs.ball.y + 1, gs.p2.paddle.x}, gs.screen)
		gs.p2.paddle.yTop = gs.ball.y - 2
		gs.p2.paddle.yBot = gs.ball.y + 1
	}

}

func drawPaddle(p *paddle, s tcell.Screen) {
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
