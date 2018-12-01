package fsm

import (
	"fmt"
	"log"
)

func (u *Updater) logf(f string, a ...interface{}) {
	f = "%s " + f
	a2 := []interface{}{fmt.Sprintf("updater(%s)", u.C.ID)}
	a2 = append(a2, a...)
	log.Printf(f, a2...)
}

func (u *Updater) debugf(f string, a ...interface{}) {
	if u.debug {
		u.logf(f, a...)
	}
}

func (u *Updater) SetDebug(debug bool) {
	u.debug = debug
}
