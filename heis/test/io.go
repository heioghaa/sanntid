package bla

import (
	"../liftio"
)
func main() {
	order := make (chan uint)
	status := make(chan liftio.FloorStatus)
	Init(order, status)
	return 1
}
