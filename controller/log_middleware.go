package controller

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (s *Controller) LogsMiddleware(c *gin.Context) {
	ID := c.Param("ID")

	s.mu.Lock()
	wsConn, ok := s.wsMap[ID]
	s.mu.Unlock()
	if !ok {
		return
	}

	// request logs
	logs := "request: \n"
	logs += c.Request.Method + " " + c.Request.RequestURI + " " + c.Request.Proto + "\n"
	for k, v := range c.Request.Header {
		logs += k + ": " + v[0] + "\n"
	}
	logs += "\n"

	var buf bytes.Buffer
	tee := io.TeeReader(c.Request.Body, &buf)
	body, _ := ioutil.ReadAll(tee)
	logs += string(body) + "\n"
	c.Request.Body = ioutil.NopCloser(&buf)
	log.Infof(logs)

	wsConn.ch <- logs

	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()

	// response log

	logs = "response: \n"
	logs += c.Request.Proto + " " + strconv.Itoa(c.Writer.Status()) + " " +
		http.StatusText(c.Writer.Status()) + "\n"

	for k, v := range c.Writer.Header() {
		logs += k + ": " + v[0] + "\n"
	}
	logs += "\n"

	ct := blw.Header().Get("Content-Type")
	if strings.Contains(ct, "json") || strings.Contains(ct, "form") || ct == "" {
		logs += blw.body.String() + "\n"
	} else {
		logs += "... html response ...\n"
	}

	wsConn.ch <- logs
}
