package game

import (
	"fmt"
	"time"
	
	"sdl"
)

const (
	exitFile = "assets/door.png"
	exitFrames = 4
	exitFrameWidth, exitFrameHeight = 64, 64
)

type DoorState int
const (
	DoorStateClosed DoorState = iota
	DoorStateOpen
	DoorStateQuit
)

type Exit struct {
	tex *sdl.Texture
	DoorState
	x, y, frame int
	
	inbox chan message
	Updater *time.Ticker
	Controller chan DoorState
}

func NewExit(ctx *sdl.Context) (*Exit, error) {
	tex, err := ctx.GetTexture(exitFile)
	if err != nil {
		return nil, err
	}
	e := &Exit{
		tex: tex,
		inbox: make(chan message, 10),
		Controller: make(chan DoorState, 10),
		Updater: time.NewTicker(100 * time.Millisecond),
	}
	kmp("player.location", e.inbox)
	go e.life()
	return e, nil
}

func (e *Exit) AddChild(Object) {}
func (e *Exit) Children() []Object {return nil}

func (e *Exit) Destroy() {
	e.Controller <- DoorStateQuit
	close(e.Controller)
	e.Updater.Stop()
}

func (e *Exit) Draw(r *sdl.Renderer) error {
	return r.Copy(e.tex, sdl.Rect(e.frame*exitFrameWidth, 0, exitFrameWidth, exitFrameHeight), sdl.Rect(e.x, e.y, exitFrameWidth, exitFrameHeight))
}

func (e *Exit) String() string {
	return fmt.Sprintf("%T", e)
}

func (e* Exit) life() {
	for {
		select {
		case msg := <-e.inbox:
			switch m := msg.v.(type) {
			case locationMsg:
				if msg.k == "player.location" {
					// If the player is near the door, open it;
					// If the player is not near the door, close it.
					if e.DoorState == DoorStateClosed &&
						m.x > e.x - 200 && m.x < e.x + 200 &&
						m.y > e.y - 200 && m.y < e.y + 200 {
						e.DoorState = DoorStateOpen
					}
					if e.DoorState == DoorStateOpen && (
						m.x <= e.x - 200 || m.x >= e.x + 200 ||
						m.y <= e.y - 200 || m.y >= e.y + 200) {
						e.DoorState = DoorStateClosed
					}
				}
			}
		case c := <-e.Controller:
			if c == DoorStateQuit {
				return
			}
			e.DoorState = c
		case <-e.Updater.C:
			switch  {
			case e.DoorState == DoorStateClosed && e.frame > 0:
				e.frame--
			case e.DoorState == DoorStateOpen && e.frame < 3:
				e.frame++
			}
		}
	}
}