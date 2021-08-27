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
	// map ID -> wsConn
	wsMap map[string]*wsConn
	// protect wsMap
	mu sync.Mutex

	jwtSecret string
}

// NewController creates a new controller
func NewController() *Controller {
	return &Controller{
		wsMap: make(map[string]*wsConn),
		mu:    sync.Mutex{},
		jwtSecret: "ZGQyMWE4ZDEyYzNiZGRhYWIwYzk1NzNmMjkyNDY4OTIgIG1vY2tvYXV0aAo",
	}
}
