package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/nlacasse/monome"
)

func runDevice(g *monome.Grid) {
	for {
		select {
		case keyEv := <-g.Ev:
			log.Printf("GOT KEY %x", keyEv)
			if keyEv.T == monome.KeyUp {
				continue
			}
			g.SetLED(keyEv.X, keyEv.Y, true)
			sound := fmt.Sprintf("/home/nlacasse/sounds/%d%d.wav", keyEv.X, keyEv.Y)
			c := exec.Command("mplayer", sound)
			go func(keyEv monome.KeyEv) {
				if err := c.Run(); err != nil {
					log.Print(err)
				}
				g.SetLED(keyEv.X, keyEv.Y, false)
			}(keyEv)
		case <-g.Disconnect:
			log.Printf("Disconnect!!")
			return
		}
	}
}

func main() {
	m := monome.New()

	for {
		g := <-m.Devices
		go runDevice(g)
	}
}
