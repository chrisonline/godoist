package tui

import (
	"strings"
	"testing"

	"github.com/biomassa/godoist/internal/todoist"
)

// x starts the fireworks. The show plays over the screen and ends by itself, and a tick of
// an old show does nothing.
func TestFireworksAfterComplete(t *testing.T) {
	m := openWork(t)
	m.focus = paneTasks
	m = send(t, m, key("x"))
	if m.fw == nil || len(m.fw.rockets) != fwRockets {
		t.Fatalf("no show after x: %+v", m.fw)
	}
	gen := m.fw.gen
	burst := false
	for i := 0; m.fw != nil; i++ {
		if i > 300 {
			t.Fatal("the show does not end")
		}
		if len(m.fw.sparks) > 0 && !burst {
			burst = true
			if !strings.ContainsAny(stripANSI(m.overlays(m.render())), "*+") {
				t.Error("no sparks on the screen")
			}
		}
		m = send(t, m, fireworksMsg{gen})
	}
	if !burst {
		t.Error("the rocket did not burst")
	}
	m = send(t, m, key("x"))
	if m.fw == nil || m.fw.gen == gen {
		t.Fatalf("second show: %+v", m.fw)
	}
	n := len(m.fw.rockets)
	m = send(t, m, fireworksMsg{gen})
	if m.fw.rockets[0].y != float64(m.height-2) || len(m.fw.rockets) != n {
		t.Error("a tick of the old show moved the new show")
	}
}

// A bulk completion launches more rockets, at most fwMaxRockets.
func TestFireworksBulk(t *testing.T) {
	m := openWork(t)
	var sel []*todoist.Task
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "g"} {
		sel = append(sel, &todoist.Task{ID: id, ProjectID: "p1", Content: id})
	}
	m.bulkComplete(sel)
	if m.fw == nil || len(m.fw.rockets) != fwMaxRockets {
		t.Fatalf("rockets = %+v", m.fw)
	}
}

// no_fireworks in the config turns the show off.
func TestNoFireworks(t *testing.T) {
	m := openWork(t)
	m.focus = paneTasks
	m.noFireworks = true
	m = send(t, m, key("x"))
	if m.fw != nil {
		t.Error("show with no_fireworks")
	}
}
