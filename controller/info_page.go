package controller

import (
	"encoding/base64"
	"encoding/json"
	"mockoauth/viewmodule"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
)

// InfoPage godoc
// @Summary Authorization Code Grant mock page
// @Description show logs and details of Authorization Code Grant oauth mode
// @Produce  html
// @Router /oauth2/ [get]
func (s *Controller) InfoPage(c *gin.Context) {
	log.Infof("get info page")
	clientInfo := viewmodule.ClientInfo{}
	host := c.Request.Host

	cookieID, err := c.Cookie("client_id")
	if err != nil || cookieID == "" {
		clientInfo.ID = shortuuid.New()
		c.SetCookie("client_id", clientInfo.ID, 36000, "/oauth2/", host, false, true)
	} else {
		clientInfo.ID = cookieID
	}

	oauthProviderInfo := viewmodule.OAuthProviderInfo{
		ClientID:     clientInfo.ID,
		ClientSecret: base64.RawURLEncoding.EncodeToString([]byte(clientInfo.ID)),
	}
	c.HTML(http.StatusOK, "tmpl/info.tmpl",
		gin.H{
			"title":             "OAuth2 Provider Mock Info",
			"oauthProviderInfo": oauthProviderInfo,
			"host":              host,
			"client":            clientInfo,
		})
}

// InfoPageWS godoc
// @Summary Authorization Code Grant mock page, websocket connection
// @Description send logs to client and receive requests from clients.
// @Router /oauth2/ws/:ID [get]
func (s *Controller) InfoPageWS(c *gin.Context) {
	ID := c.Param("ID")
	log.Infof("create websocket %s", ID)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("websocket upgrade error: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	wsConn := &wsConn{
		ws: ws,
		ch: make(chan string),
	}

	s.mu.Lock()
	s.wsMap[ID] = wsConn
	s.mu.Unlock()

	rch := make(chan string)
	defer close(rch)

	go func() {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Errorf("websocket read error: %s", err)
			return
		}
		rch <- string(message)
	}()

	defer func() {
		wsConn.ws.Close()
		close(wsConn.ch)
		s.mu.Lock()
		delete(s.wsMap, ID)
		s.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-wsConn.ch:
			var m viewmodule.WSMsg
			if !ok {
				log.Errorf("ws closed")
				m.Type = "closed"
				m.Data = "ws closed"
				toSend, _ := json.Marshal(&m)
				ws.WriteMessage(websocket.TextMessage, toSend)
				return
			}
			m.Type = "log"
			// m.Data = html.EscapeString(msg)
			m.Data = msg
			toSend, _ := json.Marshal(&m)
			err := ws.WriteMessage(websocket.TextMessage, toSend)
			if err != nil {
				log.Errorf("ws write error: %v", err)
				return
			}
		case msg := <-rch:
			log.Infof("get message from client: %s", msg)
		case <-time.After(time.Second*10):
			var m viewmodule.WSMsg
			m.Type = "ping"
			toSend, _ := json.Marshal(&m)
			err := ws.WriteMessage(websocket.TextMessage, toSend)
			if err != nil {
				log.Errorf("ws write error: %v", err)
				return
			}
		}
	}
}
