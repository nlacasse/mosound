package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/nlacasse/monome"
)

var (
	soundDir   = flag.String("sound-dir", "/home/nlacasse/sounds", "directory containing sound files")
	serialoscd = flag.String("serialoscd", "", "serialoscd path, which will be started if not empty")
)

func drawInit(g *monome.Grid) {
	for i := 0; i < g.Size[0]; i++ {
		for j := 0; j < g.Size[1]; j++ {
			g.SetLED(i, j, true)
			time.Sleep(25 * time.Millisecond)
		}
	}
	for i := 0; i < g.Size[0]; i++ {
		for j := 0; j < g.Size[1]; j++ {
			g.SetLED(i, j, false)
			time.Sleep(25 * time.Millisecond)
		}
	}
}

func runDevice(g *monome.Grid) {
	drawInit(g)

	// Counts number of samples playing per button.
	var stateMu sync.Mutex
	var state [][]uint
	for i := 0; i < g.Size[0]; i++ {
		state = append(state, make([]uint, g.Size[1]))
	}

	for {
		select {
		case keyEv := <-g.Ev:
			if keyEv.T == monome.KeyUp {
				continue
			}

			stateMu.Lock()
			if state[keyEv.X][keyEv.Y] == 0 {
				g.SetLED(keyEv.X, keyEv.Y, true)
			}
			state[keyEv.X][keyEv.Y]++
			stateMu.Unlock()

			sound := filepath.Join(*soundDir, fmt.Sprintf("%d%d.wav", keyEv.X, keyEv.Y))
			c := exec.Command("aplay", sound)
			go func(keyEv monome.KeyEv) {
				if err := c.Run(); err != nil {
					log.Print(err)
				}
				stateMu.Lock()
				state[keyEv.X][keyEv.Y]--
				if state[keyEv.X][keyEv.Y] == 0 {
					g.SetLED(keyEv.X, keyEv.Y, false)
				}
				stateMu.Unlock()
			}(keyEv)
		case <-g.Disconnect:
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

		go func() {
			c.Wait()
			// If we get here, serialosc has crashed.  Bad news.
			log.Panicf("serialosc has stopped!")
		}()
	}

	m := monome.New()
	for {
		g := <-m.Devices
		go runDevice(g)
	}
}
