package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	SCREEN_WIDTH  = 720
	SCREEN_HEIGHT = 480
	NUM_PARTICLES = 100
	RADIUS        = 3
	MIN_DIST      = 40
	MAX_DIST      = 80
	MIN_DIST2     = MIN_DIST * MIN_DIST
	MAX_DIST2     = MAX_DIST * MAX_DIST
	INTERP_RADIUS = 1
)

var (
	texture  *sdl.Texture
	renderer *sdl.Renderer
)

var (
	pixels    = [SCREEN_HEIGHT][SCREEN_WIDTH]vec4b{}
	particles = [NUM_PARTICLES]particle{}
)

type vec4b struct {
	b, g, r, a byte
}

type vec2 struct {
	x, y float64
}

func (v vec2) sub(a vec2) vec2 {
	return vec2{v.x - a.x, v.y - a.y}
}

func (v vec2) add(a vec2) vec2 {
	return vec2{v.x + a.x, v.y + a.y}
}

func (v vec2) mul(a vec2) vec2 {
	return vec2{v.x * a.x, v.y * a.y}
}

func (v vec2) smul(s float64) vec2 {
	return vec2{v.x * s, v.y * s}
}

func (v vec2) dot(a vec2) float64 {
	return v.x*a.x + v.y*a.y
}

func (v vec2) length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v vec2) distance(a vec2) float64 {
	return v.sub(a).length()
}

type particle struct {
	p, v vec2
}

func drawPixel(x, y int, col vec4b) {
	pixels[SCREEN_HEIGHT-y-1][x] = col
}

func blendPixel(x, y int, col vec4b, opacity float64) {
	pcol := &pixels[SCREEN_HEIGHT-y-1][x]
	iOpacity := 1 - opacity
	pcol.b = byte(math.Floor(float64(pcol.b)*iOpacity + float64(col.b)*opacity))
	pcol.g = byte(math.Floor(float64(pcol.g)*iOpacity + float64(col.g)*opacity))
	pcol.r = byte(math.Floor(float64(pcol.r)*iOpacity + float64(col.r)*opacity))
}

func lerp(v0, v1, t float64) float64 {
	return v0 + t*(v1-v0)
}

func clamp(x, min, max float64) float64 {
	return math.Max(float64(min), math.Min(float64(max), float64(x)))
}

func drawCircle(p vec2, radius float64) {
	sx := int(math.Floor(math.Max(p.x-radius-INTERP_RADIUS, 0)))
	ex := int(math.Floor(math.Min(p.x+radius+INTERP_RADIUS, SCREEN_WIDTH-1)))
	sy := int(math.Floor(math.Max(p.y-radius-INTERP_RADIUS, 0)))
	ey := int(math.Floor(math.Min(p.y+radius+INTERP_RADIUS, SCREEN_HEIGHT-1)))
	for iy := sy; iy <= ey; iy++ {
		for ix := sx; ix <= ex; ix++ {
			d := math.Sqrt((float64(ix)-p.x)*(float64(ix)-p.x) + (float64(iy)-p.y)*(float64(iy)-p.y))
			diff := d - radius
			if diff < 0 {
				drawPixel(ix, iy, vec4b{255, 255, 255, 255})
			} else if diff < INTERP_RADIUS {
				opacity := 1 - diff/INTERP_RADIUS
				blendPixel(ix, iy, vec4b{255, 255, 255, 255}, opacity)
			}
		}
	}
}

func lineDist(a, b, p vec2) float64 {
	pa, ba := p.sub(a), b.sub(a)
	h := clamp(pa.dot(ba)/ba.dot(ba), 0, 1)
	return pa.sub(ba.smul(h)).length()
}

func toInt(b bool) int {
	if b {
		return 1
	}
	return -1
}

func drawLine(a, b vec2, radius, opacity float64) {
	delta := b.sub(a)
	sx, sy := toInt(delta.x >= 0), toInt(delta.y >= 0)
	if math.Abs(delta.y) < 1 || math.Abs(delta.x) < 1 {
		return
	}
	for j := -INTERP_RADIUS - radius; j <= math.Abs(delta.x)+radius+INTERP_RADIUS; j++ {
		ix := math.Floor(a.x + j*float64(sx))
		if ix >= 0 && ix < SCREEN_WIDTH {
			for i := -INTERP_RADIUS - radius; i <= math.Abs(delta.y)+radius+INTERP_RADIUS; i++ {
				iy := math.Floor(a.y + i*float64(sy))
				if iy >= 0 && iy < SCREEN_HEIGHT {
					p := vec2{ix, iy}
					d := lineDist(a, b, p)
					d = math.Min(d, math.Min(p.distance(a), p.distance(b)))
					diff := d - radius
					if diff < 0 {
						blendPixel(int(ix), int(iy), vec4b{255, 255, 255, 255}, opacity)
					} else if diff < INTERP_RADIUS {
						popacity := opacity * (1 - diff/INTERP_RADIUS)
						blendPixel(int(ix), int(iy), vec4b{255, 255, 255, 255}, popacity)
					}
				}
			}
		}
	}
}
func clearScreen() {
	for i := range pixels {
		for j := range pixels[i] {
			pixels[i][j] = vec4b{54, 47, 41, 255}
		}
	}
}
func drawScene() {
	clearScreen()
	for i := 0; i < NUM_PARTICLES; i++ {
		p := &particles[i].p
		v := &particles[i].v
		*p = (*p).add(*v)
		if p.x < -MAX_DIST {
			p.x += SCREEN_WIDTH + MAX_DIST*2
		} else if p.x > SCREEN_WIDTH+MAX_DIST {
			p.x -= SCREEN_WIDTH + MAX_DIST*2
		}
		if p.y < -MAX_DIST {
			p.y += SCREEN_HEIGHT + MAX_DIST*2
		} else if p.y > SCREEN_HEIGHT+MAX_DIST {
			p.y -= SCREEN_HEIGHT + MAX_DIST*2
		}
		v.x += 0.05*(rand.Float64()-0.5) - 0.001*v.x
		v.y += 0.05*(rand.Float64()-0.5) - 0.001*v.y
		drawCircle(*p, RADIUS)
	}

	for i := 0; i < NUM_PARTICLES; i++ {
		pi := &particles[i]
		for j := i + 1; j < NUM_PARTICLES; j++ {
			pj := particles[j]
			d := pi.p.sub(pj.p)
			d2 := d.dot(d)
			if d2 < MAX_DIST2 {
				opacity := float64(1.0)
				if d2 > MIN_DIST2 {
					opacity = (MAX_DIST2 - d2) / (MAX_DIST2 - MIN_DIST2)
				}
				drawLine(pi.p, pj.p, 1, opacity)
			}
		}
	}
}

func pixelsToBytesSlice() (result []byte) {
	for i := 0; i < len(pixels); i++ {
		for _, v := range pixels[i] {
			result = append(result, v.b)
			result = append(result, v.g)
			result = append(result, v.r)
			result = append(result, v.a)
		}
	}
	return
}

func uploadPixels() {
	pixelsToDraw := pixelsToBytesSlice()
	texture.Update(nil, pixelsToDraw, SCREEN_WIDTH*4)
	renderer.Copy(texture, nil, nil)
	renderer.Present()
}

func draw() {
	drawScene()
	uploadPixels()
}

func initParticles() {
	for i := range particles {
		particles[i] = particle{p: vec2{rand.Float64() * float64(SCREEN_WIDTH), rand.Float64() * float64(SCREEN_HEIGHT)}, v: vec2{0.0, 0.0}}
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	initParticles()

	window, err := sdl.CreateWindow("DOTS", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, SCREEN_WIDTH, SCREEN_HEIGHT, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)

	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, SCREEN_WIDTH, SCREEN_HEIGHT)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	texture.SetBlendMode(sdl.BLENDMODE_NONE)

	lastTicks := sdl.GetTicks()
	fps := 0
	running := true
	for running {
		ticks := sdl.GetTicks()
		if ticks-lastTicks >= 1000 {
			fmt.Println("FPS:", fps)
			lastTicks = ticks
			fps = 0
		}
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				switch t.Keysym.Scancode {
				case sdl.SCANCODE_ESCAPE:
					running = false
					break
				}
			case *sdl.QuitEvent:
				running = false
				break
			}
		}
		draw()
		fps++
	}
}
