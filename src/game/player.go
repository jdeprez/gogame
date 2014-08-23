package game

import (
	//"fmt"
	"math"
	"math/rand"
	"time"

	"sdl"
)

const (
	texturePlayerFile = "assets/spacepsn.png"
	playerFramesWidth  = 4
	playerWidth, playerHeight = 32, 32
	
	playerWalkSpeed = 256 // pixels per second?
	playerJumpSpeed = -512
	playerGravity = 2048
	
	playerTau = 0.2
)

type Facing int
const (
	Left Facing = iota
	Right
)

type Animation int
const (
	Standing Animation = iota
	Walking
	Jumping
	Falling
)

type Control int
const (
	Quit Control = iota
	StartWalkLeft
	StopWalkLeft
	StartWalkRight
	StopWalkRight
	StartJump
	StopJump
	StartFire
	StopFire
	Land
	Teleport
)

type Player struct {
	Base
	tex  *sdl.Texture
	lastUpdate time.Duration
	
	Controller chan Control
	Ticker chan time.Duration
	
	// All struct elements below should be controlled only by the life goroutine.
	facing Facing
	anim Animation
	wx, wy float64
	x, y, frame int
	fx, fy, dx, dy, ddx, ddy float64
}

func NewPlayer(ctx *sdl.Context) (*Player, error) {
	tex, err := ctx.GetTexture(texturePlayerFile)
	if err != nil {
		return nil, err
	}
	p := &Player{
		fx:   float64(rand.Int31() % (1024 - playerWidth)),
		fy:   float64(rand.Int31() % (768 - playerHeight)),
		tex: tex,
		Controller: make(chan Control),
		Ticker: make(chan time.Duration),
	}
	go p.life()
	return p, nil
}

func (p *Player) Draw(r *sdl.Renderer) error {
	fx := (p.frame%playerFramesWidth)*playerWidth
	switch p.facing {
	case Left:
		return r.Copy(p.tex,
			sdl.Rect(fx, 0, playerWidth, playerHeight),
			sdl.Rect(p.x, p.y, playerWidth, playerHeight))
	case Right:
		// TODO: separate animations for facing right
		return r.CopyEx(p.tex,
			sdl.Rect(fx, 0, playerWidth, playerHeight),
			sdl.Rect(p.x, p.y, playerWidth, playerHeight), 0, nil, sdl.FlipHorizontal)
	}
	return nil
}

func (p *Player) update(t time.Duration) {
	if p.lastUpdate == 0 {
		p.lastUpdate = t
		return
	}
	delta := float64(t - p.lastUpdate) / float64(time.Second)
	
	switch p.anim {
	case Walking:
		p.frame = (int(2 * t / time.Millisecond) % 1000) / 250
		tau := playerTau * math.Exp(delta)
		p.dx = tau * p.wx + (1-tau) * p.dx
		p.dy = tau * p.wy + (1-tau) * p.dy
	case Falling:
		p.frame = 0
		p.dx += p.ddx * delta
		p.dy += p.ddy * delta
	default:
		p.frame = 0
		tau := playerTau * math.Exp(delta)
		p.dx = tau * p.wx + (1-tau) * p.dx
		p.dy = tau * p.wy + (1-tau) * p.dy
	}

	// FISIXX
	p.fx += p.dx * delta
	p.fy += p.dy * delta
	p.x = int(p.fx)
	p.y = int(p.fy)
	p.lastUpdate = t
}

func (p *Player) life() {
	for {
		select {
		case ctl := <-p.Controller:
			switch ctl {
			case Quit:
				return
			case StartWalkLeft:
				switch p.anim {
				case Standing, Walking:
					p.anim = Walking
					p.facing = Left
					p.wx = -playerWalkSpeed
				}
			case StopWalkLeft:
				if p.anim == Walking {
					p.anim = Standing
					p.wx = 0
				}
			case StartWalkRight:
				switch p.anim {
				case Standing, Walking:
					p.anim = Walking
					p.facing = Right
					p.wx = playerWalkSpeed
				}
			case StopWalkRight:
				if p.anim == Walking {
					p.anim = Standing
					p.wx = 0
				}
			case StartJump:
				switch p.anim {
				case Standing, Walking:
					p.anim = Jumping
				}
			case StopJump:
				if p.anim == Jumping {
					p.anim = Falling
					p.dy = playerJumpSpeed
					p.ddy = playerGravity
				}
			case Land:
				if p.anim == Falling {
					p.anim = Standing
					p.dy = 0
					p.ddy = 0
				}
			default:
				// TODO: more actions
			}
		case t := <-p.Ticker:
			p.update(t)
		}
	}
}