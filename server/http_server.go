package server

import (
	"mockoauth/middleware"
	"mockoauth/resources"
	"mockoauth/router"

	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	c *gin.Engine

	// listening on address, ":8080" for example
	ListenOn string
}

func NewHTTPServer() *HTTPServer {
	s := &HTTPServer{
		c: gin.New(),
	}

	// set resources and template before any router
	resources.Init(s.c)

	middleware.Init(s.c)
	router.Init(s.c)

	return s
}

func (s *HTTPServer) Start() {
	s.c.Run(s.ListenOn)
}

func (s *HTTPServer) Router() *gin.Engine {
	return s.c
}
