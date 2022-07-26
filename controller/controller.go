package controller

import (
	"sync"

	"github.com/gorilla/websocket"
)

// wsConn wrap websocket connection with session info
type wsConn struct {
	ws *websocket.Conn

	// ws handler wait on channel to receive messages
	ch chan string
}

// Controller is a oauth mock controller
type Controller struct {
	// map ID -> []wsConn
	wsProviderMap WSMap
	wsClientMap WSMap

	jwtSecret string
}

// NewController creates a new controller
func NewController() *Controller {
	r := &Controller{
		jwtSecret: "ZGQyMWE4ZDEyYzNiZGRhYWIwYzk1NzNmMjkyNDY4OTIgIG1vY2tvYXV0aAo",
	}
	r.wsClientMap.Init()
	r.wsProviderMap.Init()
	return r
}

type WSMap struct {
	// ID -> []wsConn
	m map[string][]*wsConn
	mu sync.Mutex
}

func (m* WSMap) Init() {
	m.m = make(map[string][]*wsConn)
}

func (m *WSMap) Put(id string, conn *wsConn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.m[id] = append(m.m[id], conn)
}

func (m *WSMap) Remove(id string, conn *wsConn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clist := m.m[id]
	var newList []*wsConn
	for _, c := range clist {
		if c != conn {
			newList = append(newList, c)
		}
	}
	m.m[id] = newList
}

func (m *WSMap) Get(id string) []*wsConn {
	m.mu.Lock()
	defer m.mu.Unlock()

	clist := m.m[id]

	return clist
}

func (m *WSMap) SendTo(id string, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clist := m.m[id]
	for _, c := range clist {
		c.ch <- msg
	}
}
