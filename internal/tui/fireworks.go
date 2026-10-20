package tui

import (
	"math"
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// fireworks is the show that celebrates a completion: rockets rise from the bottom of the
// screen and burst into sparks that fall and fade. It plays over the screen and does not
// take keys, so you can work on while it plays.
type fireworks struct {
	gen     int // the show of the tick messages, so that an old tick does not run a new show
	rockets []rocket
	sparks  []spark
}

type rocket struct {
	x, y, vy  float64
	wait      int    // frames before the launch
	hex, hex2 string // the colors of the burst
	kind      burstKind
}

type burstKind int

const (
	burstRing    burstKind = iota // one ring in one color
	burstTwoTone                  // an outer ring and an inner ring in another color
	burstGlitter                  // a ring with sparks that twinkle
	burstWillow                   // slow sparks that hang long and fall
	burstKinds
)

type spark struct {
	x, y, vx, vy float64
	life, total  int // frames left, frames at the burst
	hex          string
	twinkle      bool
}

const (
	fwFrame      = 33 * time.Millisecond
	fwRocketG    = 0.08 // the gravity of a rocket, in rows per frame²
	fwSparkG     = 0.02 // the gravity of a spark
	fwDrag       = 0.92 // the speed that a spark keeps from one frame to the next
	fwRockets    = 5    // the rockets for one completed task
	fwMaxRockets = 12   // the rockets of one show at most
)

// fwColors are the burst colors. The accent of the theme joins them.
var fwColors = []string{"#FF5E5B", "#FFD93D", "#6BCB77", "#4D96FF", "#C77DFF", "#FF9F1C"}

type fireworksMsg struct{ gen int }

func fireworksTick(gen int) tea.Cmd {
	return tea.Tick(fwFrame, func(time.Time) tea.Msg { return fireworksMsg{gen} })
}

// celebrate launches a volley of fwRockets rockets, and two more for each further completed
// task of n, at most fwMaxRockets. A show that plays gets the new rockets.
func (m *Model) celebrate(n int) tea.Cmd {
	if m.noFireworks || m.width < 20 || m.height < 8 {
		return nil
	}
	n = min(fwRockets+2*(max(n, 1)-1), fwMaxRockets)
	start := m.fw == nil
	if start {
		m.fwGen++
		m.fw = &fireworks{gen: m.fwGen}
	}
	wait := 0
	for range n {
		m.fw.rockets = append(m.fw.rockets, newRocket(m.width, m.height, wait))
		wait += 2 + rand.IntN(7)
	}
	if !start {
		return nil
	}
	return fireworksTick(m.fw.gen)
}

// newRocket starts at the bottom and bursts in the upper half of the screen.
func newRocket(w, h, wait int) rocket {
	y := float64(h - 2)
	top := float64(h) * (0.12 + 0.3*rand.Float64())
	colors := append(fwColors[:len(fwColors):len(fwColors)], hexAccent)
	return rocket{
		x:    float64(w) * (0.1 + 0.8*rand.Float64()),
		y:    y,
		vy:   -math.Sqrt(2 * fwRocketG * (y - top)),
		wait: wait,
		hex:  colors[rand.IntN(len(colors))],
		hex2: colors[rand.IntN(len(colors))],
		kind: burstKind(rand.IntN(int(burstKinds))),
	}
}

// fireworksFrame moves the show one frame on, and ends it when nothing is left.
func (m Model) fireworksFrame(msg fireworksMsg) (tea.Model, tea.Cmd) {
	if m.fw == nil || msg.gen != m.fw.gen {
		return m, nil
	}
	if !m.fw.step(m.height) {
		m.fw = nil
		return m, nil
	}
	return m, fireworksTick(msg.gen)
}

// step moves the rockets and the sparks. It reports whether anything is left.
func (f *fireworks) step(h int) bool {
	rs := f.rockets[:0]
	for _, r := range f.rockets {
		if r.wait > 0 {
			r.wait--
			rs = append(rs, r)
			continue
		}
		r.y += r.vy
		r.vy += fwRocketG
		if r.vy >= 0 { // the top of the flight
			f.burst(r, h)
			continue
		}
		rs = append(rs, r)
	}
	f.rockets = rs
	ss := f.sparks[:0]
	for _, s := range f.sparks {
		if s.life--; s.life <= 0 {
			continue
		}
		s.x += s.vx
		s.y += s.vy
		s.vx *= fwDrag
		s.vy = s.vy*fwDrag + fwSparkG
		ss = append(ss, s)
	}
	f.sparks = ss
	return len(f.rockets)+len(f.sparks) > 0
}

// burst puts the sparks of the rocket's kind at the rocket.
func (f *fireworks) burst(r rocket, h int) {
	scale := min(1, float64(h)/30)
	switch r.kind {
	case burstTwoTone:
		f.ring(r.x, r.y, 30, 0.6*scale, 18, r.hex, false)
		f.ring(r.x, r.y, 16, 0.3*scale, 14, r.hex2, false)
	case burstGlitter:
		f.ring(r.x, r.y, 36, 0.55*scale, 26, r.hex, true)
	case burstWillow:
		f.ring(r.x, r.y, 28, 0.35*scale, 45, r.hex, false)
	default:
		f.ring(r.x, r.y, 32, 0.55*scale, 20, r.hex, false)
	}
}

// ring puts n sparks around x, y with about speed v and life frames. A terminal cell is
// about twice as high as it is wide, so the sparks move twice as fast across as up and down.
func (f *fireworks) ring(x, y float64, n int, v float64, life int, hex string, twinkle bool) {
	for i := range n {
		a := 2*math.Pi*float64(i)/float64(n) + rand.Float64()*0.3
		sv := v * (0.75 + 0.5*rand.Float64())
		l := life + rand.IntN(life/2+1)
		f.sparks = append(f.sparks, spark{x: x, y: y, vx: 2 * sv * math.Cos(a), vy: sv * math.Sin(a),
			life: l, total: l, hex: hex, twinkle: twinkle})
	}
}

// layers gives one layer for each rocket and spark on a w × h screen, above all dialogs.
func (f *fireworks) layers(w, h int) []*lipgloss.Layer {
	var ls []*lipgloss.Layer
	put := func(x, y float64, s lipgloss.Style, ch string) {
		cx, cy := int(math.Round(x)), int(math.Round(y))
		if cx < 0 || cy < 0 || cx >= w || cy >= h {
			return
		}
		ls = append(ls, lipgloss.NewLayer(s.Render(ch)).X(cx).Y(cy).Z(6))
	}
	for _, r := range f.rockets {
		if r.wait > 0 {
			continue
		}
		put(r.x, r.y+1, st(r.hex).Faint(true), "·")
		put(r.x, r.y, st(r.hex).Bold(true), "•")
	}
	for _, s := range f.sparks {
		if s.twinkle { // a glitter spark blinks off or flashes white
			switch rand.IntN(3) {
			case 0:
				continue
			case 1:
				put(s.x, s.y, st("#FFFFFF").Bold(true), "*")
				continue
			}
		}
		switch frac := float64(s.life) / float64(s.total); {
		case frac > 0.66:
			put(s.x, s.y, st(s.hex).Bold(true), "*")
		case frac > 0.33:
			put(s.x, s.y, st(s.hex), "+")
		default:
			put(s.x, s.y, st(s.hex).Faint(true), "·")
		}
	}
	return ls
}
