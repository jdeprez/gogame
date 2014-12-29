// Package game contains the game logic and stuff.
package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"sdl"
)

const (
	clockDuration = 10 * time.Millisecond

	level0File  = "assets/level0.txt"
	level1AFile = "assets/level1a.txt"
	level1BFile = "assets/level1b.txt"
)

type gameState int

const (
	gameStateRunning = iota
	gameStateMenu
	gameStateQuitting
)

var gameInstance *Game

type Game struct {
	state gameState
	ctx   *sdl.Context
	clock *time.Ticker
	inbox chan message

	renderer   *sdl.Renderer
	wv         *worldView
	world, hud complexBase

	cursor *sprite
	menu   *menu
}

func ctx() *sdl.Context {
	return gameInstance.ctx
}

func NewGame(ctx *sdl.Context) (*Game, error) {
	if gameInstance != nil {
		return gameInstance, nil
	}

	g := &Game{
		state:    gameStateMenu,
		ctx:      ctx,
		clock:    time.NewTicker(clockDuration),
		renderer: ctx.Renderer,
		wv: &worldView{
			view:  sdl.Rect{0, 0, 1024, 768},
			world: sdl.Rect{0, 0, 4096, 768}, // TODO: derive from terrain
		},
		inbox:  make(chan message, 10),
		cursor: &sprite{TemplateKey: "cursor"},
	}
	gameInstance = g

	// Attach cursor to events.
	go cursorLife(g.cursor)

	menu, err := newMenu()
	if err != nil {
		return nil, err
	}
	g.menu = menu

	// Test hexagons...
	for i := 0; i < 25; i++ {
		for j := 0; j < 8; j++ {
			g.world.addChild(&sprite{
				TemplateKey: "hex",
				X:           192*j + 96*(i%2) - 32,
				Y:           -rand.Intn(5) * 2,
				Z:           32 * (i - 1),
			})
		}
	}

	g.world.addChild(&orb{X: 150, Y: 22, Z: 100, Selected: true})

	kmp("quit", g.inbox)
	kmp("player.location", g.inbox)
	kmp("input.event", g.inbox)
	kmp("menuAction", g.inbox)
	go g.life()
	go g.pulse()
	return g, nil
}

func (g *Game) life() {
	defer func() {
		g.state = gameStateQuitting
	}()
	for msg := range g.inbox {
		//log.Printf("game.inbox got %+v\n", msg)
		switch msg.k {
		case "quit":
			return
		case "menuAction":
			switch msg.v.(string) {
			case "start":
				g.menu.Invisible = true
				g.state = gameStateRunning
			case "levelEdit":
				g.menu.Invisible = true
			}
		}
		switch m := msg.v.(type) {
		case locationMsg:
			if msg.k == "player.location" {
				g.wv.focus(m.x, m.y)
			}
		case sdl.QuitEvent:
			quit()
		case *sdl.KeyUpEvent:
			switch m.KeyCode {
			case 'q':
				quit()
			case 'e':
				// Do teleport
				//g.lev.active = (g.lev.active + 1) % 2
			}
		}
	}
}

// pulse notifies the "clock" key with events from g.clock when the game is
// playing.
func (g *Game) pulse() {
	t0 := time.Now()
	for t := range g.clock.C {
		if g.state == gameStateRunning {
			notify("clock", t.Sub(t0))
		}
	}
}

func (g *Game) Draw() error {
	// Draw everything in the world in world coordinates.
	g.renderer.PushOffset(-g.wv.view.X, -g.wv.view.Y)
	defer g.renderer.ResetOffset()
	if err := g.world.draw(g.renderer); err != nil {
		return err
	}
	g.renderer.PopOffset()

	// Draw the HUD in screen coordinates.
	if err := g.hud.draw(g.renderer); err != nil {
		return err
	}

	// Draw the menu in screen coordinates.
	if err := g.menu.draw(g.renderer); err != nil {
		return err
	}

	// Draw the cursor, always, in screen coordinates.
	return g.cursor.draw(g.renderer)
}

func (g *Game) Destroy() {
	//log.Print("game.destroy")
	g.clock.Stop()
	g.world.destroy()
	g.hud.destroy()
}

func (g *Game) Exec(cmd string) {
	//log.Printf("game.Exec(%q)\n", cmd)
	argv := strings.Split(cmd, " ")
	switch argv[0] {
	case "quit":
		quit()
	case "help":
		if len(argv) == 1 {
			fmt.Println("help: Usage: help <command>")
		} else {
			fmt.Println("help: Not yet implemented")
		}
	case "":
		return
	default:
		fmt.Println("Bad command or file name")
	}
}

func (g *Game) Quitting() bool {
	return g.state == gameStateQuitting
}

/*
func (g *Game) level() *level {
	return g.levels[g.lev.active]
}
*/

func (g *Game) HandleEvent(ev sdl.Event) error {
	notify("input.event", ev)
	return nil
}

type worldView struct {
	view, world sdl.Rect
}

// focus moves the world viewport to include the point. Generally this
// would be used to focus on the player. It snaps immediately, no smoothing.
func (r *worldView) focus(x, y int) {
	// Keep the point in view.
	left, right := r.view.W/4, 3*r.view.W/4
	if x-r.view.X > right {
		r.view.X = x - right
	}
	if x-r.view.X < left {
		r.view.X = x - left
	}
	// Clamp to world bounds.
	if r.view.X < r.world.X {
		r.view.X = r.world.X
	}
	if r.view.X+r.view.W > r.world.X+r.world.W {
		r.view.X = r.world.X + r.world.W - r.view.W
	}

	top, bottom := r.view.H/4, 3*r.view.H/4
	if y-r.view.Y < top {
		r.view.Y = y - top
	}
	if y-r.view.Y > bottom {
		r.view.Y = y - bottom
	}
	if r.view.Y < r.world.Y {
		r.view.Y = r.world.Y
	}
	if r.view.Y+r.view.H > r.world.Y+r.world.H {
		r.view.Y = r.world.Y + r.world.H - r.view.H
	}
}
