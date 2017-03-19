package monome

import (
	"fmt"
	"log"
	"sync"

	"github.com/nlacasse/go-osc/osc"
)

const (
	sport = 45450
	cport = 12002
)

type Monome struct {
	s *osc.Server
	c *osc.Client

	mu        sync.Mutex
	deviceMap map[string]*Grid

	Devices chan *Grid
}

func New() *Monome {
	m := &Monome{
		s:         &osc.Server{Addr: fmt.Sprintf("127.0.0.1:%d", sport)},
		c:         osc.NewClient("127.0.0.1", cport),
		deviceMap: make(map[string]*Grid),
		Devices:   make(chan *Grid),
	}
	m.s.Handle("/serialosc/device", m.handleDevice)
	m.s.Handle("/serialosc/add", m.handleAdd)
	m.s.Handle("/serialosc/remove", m.handleRemove)

	go func() {
		if err := m.s.ListenAndServe(); err != nil {
			log.Panic(err)
		}
	}()

	m.sendNotify()
	m.sendList()

	return m
}

func (m *Monome) newDevice(id string, typ string, port int32) {
	// TODO: Support arc devices.  Will need interface type to abstract Grid
	// and Arc types.
	if typ != "monome 40h" {
		log.Panicf("unknown device type: %q", typ)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.deviceMap[id]; ok {
		return
	}

	g := NewGrid(port)
	// Wait for g to be ready.
	<-g.ready

	// Send g!
	m.deviceMap[id] = g
	m.Devices <- g
}

func (m *Monome) sendList() {
	log.Printf("sendList")
	m.c.Send(osc.NewMessage("/serialosc/list", "127.0.0.1", int32(sport)))
}

func (m *Monome) sendNotify() {
	log.Printf("sendNotify")
	m.c.Send(osc.NewMessage("/serialosc/notify", "127.0.0.1", int32(sport)))
}

func (m *Monome) handleDevice(msg *osc.Message) {
	log.Printf("handleDevice: %v", msg)
	id := msg.Arguments[0].(string)
	typ := msg.Arguments[1].(string)
	port := msg.Arguments[2].(int32)
	m.newDevice(id, typ, port)
}

func (m *Monome) handleAdd(msg *osc.Message) {
	log.Printf("handleAdd: %v", msg)
	id := msg.Arguments[0].(string)
	typ := msg.Arguments[1].(string)
	port := msg.Arguments[2].(int32)
	m.newDevice(id, typ, port)

	// Wait for another one.
	m.sendNotify()
}

func (m *Monome) handleRemove(msg *osc.Message) {
	log.Printf("handleRemove: %+v", msg)
	id := msg.Arguments[0].(string)
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.deviceMap[id]; !ok {
		return
	}

	delete(m.deviceMap, id)

	// Wait for another one.
	m.sendNotify()
}
