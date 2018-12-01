package starlight

import (
	"fmt"
	"log"
)

func (g *Agent) logf(f string, a ...interface{}) {
	f = "%s " + f
	name := g.name
	if name == "" {
		name = fmt.Sprintf("agent %p", g)
	}
	a2 := []interface{}{name}
	a2 = append(a2, a...)
	log.Printf(f, a2...)
}

func (g *Agent) debugf(f string, a ...interface{}) {
	if g.debug {
		g.logf(f, a...)
	}
}

func (g *Agent) SetDebug(debug bool, name string) {
	g.debug = debug
	g.name = name
}
