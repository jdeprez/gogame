package game

import (
	"fmt"
	"sdl"
)

const (
	defaultFontFile = "assets/munro.ttf"
	defaultFontSize = 20
)

var defaultFont *sdl.Font

type text struct {
	t          string
	tex        *sdl.Texture
	x, y, w, h int
}

func newText(ctx *sdl.Context, s string, c sdl.Colour) (*text, error) {
	if defaultFont == nil {
		f, err := sdl.LoadFont(defaultFontFile, defaultFontSize)
		if err != nil {
			return nil, err
		}
		defaultFont = f
	}
	surf, err := defaultFont.RenderSolid(s, c)
	if err != nil {
		return nil, err
	}
	w, h := surf.Size()
	tex, err := ctx.Renderer.TextureFromSurface(surf)
	if err != nil {
		return nil, err
	}
	return &text{
		t:   s,
		tex: tex,
		w:   w,
		h:   h,
	}, nil
}

func (t *text) draw(r renderer) error {
	return r.Copy(t.tex,
		sdl.Rect{0, 0, t.w, t.h},
		sdl.Rect{t.x, t.y, t.w, t.h})
}

func (t *text) destroy() {
	if t.tex != nil {
		t.tex.Destroy()
	}
}

func (t *text) String() string {
	return fmt.Sprintf("text: %q", t.t)
}
