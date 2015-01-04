package game

import (
	"encoding/gob"
	"sdl"
	"sync"
)

func init() {
	gob.Register(&sprite{})
}

type spriteTemplate struct {
	baseX, baseY            int // point within the sprite frame of the "centre"
	framesX, framesY        int // number of frames along X, Y axes
	frameWidth, frameHeight int // size of a frame
	sheetFile               string

	tex    *sdl.Texture
	loadMu sync.Once
}

func (t *spriteTemplate) load() (err error) {
	t.loadMu.Do(func() {
		t.tex, err = ctx().LoadImage(t.sheetFile)
	})
	return
}

type sprite struct {
	Invisible      bool
	TemplateKey    string
	X, Y, Z, Frame int

	template *spriteTemplate
	w, h     int
}

func (s *sprite) bounds() sdl.Rect {
	if s == nil || s.template == nil {
		return sdl.Rect{}
	}
	return sdl.Rect{X: s.X, Y: s.Y + s.Z, W: s.w, H: s.h}
}

func (s *sprite) invisible() bool {
	return s == nil || s.Invisible
}

func (s *sprite) load() error {
	if s.template != nil {
		return nil
	}
	s.template = templateLibrary[s.TemplateKey]
	s.w, s.h = s.template.frameWidth, s.template.frameHeight
	return s.template.load()
}

func (s *sprite) draw(r *sdl.Renderer) error {
	if s == nil || s.Invisible {
		return nil
	}
	if err := s.load(); err != nil {
		return err
	}

	// Compute the frame bounds and draw.
	srcX := (s.Frame % s.template.framesX) * s.template.frameWidth
	srcY := ((s.Frame / s.template.framesX) % s.template.framesY) * s.template.frameHeight
	return r.Copy(s.template.tex,
		sdl.Rect{srcX, srcY, s.template.frameWidth, s.template.frameHeight},
		sdl.Rect{s.X - s.template.baseX, s.Y + s.Z - s.template.baseY, s.w, s.h})
}

func (s *sprite) z() int {
	return s.Z
}
