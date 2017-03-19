package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nlacasse/monome"
)

var (
	soundDir   = flag.String("sound-dir", "/home/nlacasse/sounds", "directory containing sound files")
	serialoscd = flag.String("serialoscd", "", "serialoscd path, which will be started if not empty")
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
			sound := filepath.Join(*soundDir, fmt.Sprintf("%d%d.wav", keyEv.X, keyEv.Y))
			c := exec.Command("aplay", sound)
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
	flag.Parse()
	if *serialoscd != "" {
		c := exec.Command(*serialoscd)
		if err := c.Start(); err != nil {
			log.Panicf("failed to start serialoscd binary %q: %v", *serialoscd, err)
		}
		defer c.Process.Kill()
		time.Sleep(2 * time.Second)
	}

	m := monome.New()
	for {
		g := <-m.Devices
		go runDevice(g)
	}
}
