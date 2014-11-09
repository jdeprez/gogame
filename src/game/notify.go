package game

import (
	"fmt"
)

// message is the interface for notification messages.
type message interface {
	String() string
}

// notes keeps track of all registered channels.
var notes = map[string][](chan message) {}

// kmp stands for "keep me posted", and registers a callback channel
// for messages sent to a given key.
func kmp(key string, ch chan message) {
	notes[key] = append(notes[key], ch)
}

// notify sends a message to every channel registered for a key.
func notify(key string, m message) {
	for _, n := range notes[key] {
		n <- m
	}
}

// Message types.

type basicMsg string

func (m basicMsg) String() string {
	return string(m)
}

var (
	quitMsg = basicMsg("quit")
)

type positionMsg struct {
	Obj message
	WorldX, WorldY int
}

func (p positionMsg) String() string {
	return fmt.Sprintf("Object:%v WorldX:%d WorldY:%d", p.Obj, p.WorldX, p.WorldY)
}