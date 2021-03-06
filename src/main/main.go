// Press E To Teleport (working title) - game(ish) made for Ludum Dare 30
// Programmer, etc: @DrJosh9000
package main

import (
	"errors"
	"runtime"

	"game"
	"sdl"
)

const (
	defaultWidth, defaultHeight = 1024, 768
	gameName                    = "Press E to Teleport"
)

var (
	quitting = errors.New("quitting")
)

func main() {
	// Must do rendering from the main thread, duh.
	runtime.LockOSThread()

	ctx, err := sdl.NewContext(gameName, defaultWidth, defaultHeight)
	if err != nil {
		panic(err)
	}
	defer ctx.Close()
	r := ctx.Renderer

	g, err := game.NewGame(ctx)
	if err != nil {
		panic(err)
	}
	defer g.Destroy()

	for {
		err := sdl.HandleEvents(func(e sdl.Event) error {
			switch v := e.(type) {
			case sdl.QuitEvent:
				return quitting
			case *sdl.KeyUpEvent:
				if v.KeyCode == 'q' {
					return quitting
				}
			}
			// Get the game to handle all other keys
			return g.HandleEvent(e)
		})
		if err == quitting {
			return
		}
		if err := r.Clear(); err != nil {
			panic(err)
		}
		if err := g.Draw(); err != nil {
			panic(err)
		}
		r.Present()
		sdl.Delay(1)
	}
}
