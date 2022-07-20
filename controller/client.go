package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mockoauth/viewmodule"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
)

// ClientPage godoc
// @Summary Authorization Code Grant mock page
// @Description show logs and details of Authorization Code Grant oauth mode
// @Produce  html
// @Router /oauth2/oauth2client/ [get]
func (s *Controller) ClientPage(c *gin.Context) {
	cookieID, err := c.Cookie("client_id")
	clientInfo := viewmodule.ClientInfo{}
	host := c.Request.Host
	if err != nil || cookieID == "" {
		clientInfo.ID = shortuuid.New()
		c.SetCookie("client_id", clientInfo.ID, 36000, "/oauth2client/", host, false, true)
	} else {
		clientInfo.ID = cookieID
	}

	c.HTML(http.StatusOK, "tmpl/client.tmpl",
		gin.H{
			"title":  "OAuth2 Client Mock",
			"ID":     clientInfo.ID,
			"host":   host,
			"client": clientInfo,
		})

}

// ClientPage godoc
// @Summary Authorization Code Grant mock page
// @Description show logs and details of Authorization Code Grant oauth mode
// @Produce  html
// @Router /oauth2/oauth2client/cb/:ID [get]
func (s *Controller) ClientPageCB(c *gin.Context) {
	ID := c.Param("ID")
	s.mu.Lock()
	wsConn, ok := s.wsMap[ID]
	s.mu.Unlock()

	code := c.Query("code")
	state := c.Query("state")
	log.Infof("ID=%s, code=%s, state=%s", ID, code, state)

	if ok {
		m := map[string]string{
			"code":  code,
			"state": state,
		}
		msgs, _ := json.Marshal(m)
		log.Infof("sending --code-- to rch")
		wsConn.ch <- "--code--" + string(msgs)
		log.Infof("sending --code-- ok")
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok, you can close this window and continue now."})
}

// @Router /oauth2/oauth2client/ws/:ID [get]
func (s *Controller) ClientPageWS(c *gin.Context) {
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
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Errorf("websocket read error: %s", err)
				return
			}
			log.Infof("get message from c: %s", string(message))
			rch <- string(message)
		}
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

			log.Infof("get msg from ch: %s", msg)
			if strings.HasPrefix(msg, "--code--") {
				msg = strings.TrimPrefix(msg, "--code--")
				var m map[string]string
				json.Unmarshal([]byte(msg), &m)
				log.Infof("code=%s, state=%s", m["code"], m["state"])
				m["type"] = "code"
				toSend, _ := json.Marshal(&m)
				ws.WriteMessage(websocket.TextMessage, toSend)
				continue
			}

			m.Type = "log"
			m.Data = msg
			toSend, _ := json.Marshal(&m)
			err := ws.WriteMessage(websocket.TextMessage, toSend)
			if err != nil {
				log.Errorf("ws write error: %v", err)
				return
			}
		case msg, ok := <-rch:
			if !ok {
				log.Errorf("ws closed")
				return
			}
			m := make(map[string]string)
			json.Unmarshal([]byte(msg), &m)
			if m["type"] == "pong" {
				continue
			}
			log.Infof("get message from client: %s", msg)

			values := url.Values{}
			if m["type"] == "get_token" {
				if m["grant_type"] == "authorization_code" {
					values = url.Values{
						"grant_type":    {"authorization_code"},
						"code":          {m["code"]},
						"redirect_uri":  {m["redirect_uri"]},
						"client_id":     {m["client_id"]},
						"client_secret": {m["client_secret"]},
						"scope":         {m["scope"]},
					}
				} else if m["grant_type"] == "password" {
					values = url.Values{
						"grant_type":    {"password"},
						"client_id":     {m["client_id"]},
						"client_secret": {m["client_secret"]},
						"username":      {m["username"]},
						"password":      {m["password"]},
						"scope":         {m["scope"]},
					}
				} else if m["grant_type"] == "client_credentials" {
					values = url.Values{
						"grant_type":    {"client_credentials"},
						"client_id":     {m["client_id"]},
						"client_secret": {m["client_secret"]},
						"scope":         {m["scope"]},
					}
				}
			} else if m["type"] == "refresh_token" {
				values = url.Values{
					"grant_type":    {"refresh_token"},
					"refresh_token": {m["refresh_token"]},
					"client_id":     {m["client_id"]},
					"client_secret": {m["client_secret"]},
				}
			}

			req, _ := http.NewRequest("POST", m["token_url"], strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization",
				"Basic "+base64.StdEncoding.EncodeToString(
					[]byte(m["client_id"]+":"+m["client_secret"])))

			var msCurl viewmodule.WSMsg
			msCurl.Type = "log"
			msCurl.Data = fmt.Sprintf(`>>
curl -XPOST -H "Content-Type: %s" -H "Authorization: %s" %s -d "%s"`,
				req.Header.Get("Content-Type"), req.Header.Get("Authorization"),
				m["token_url"], values.Encode())
			toSend1, _ := json.Marshal(&msCurl)
			ws.WriteMessage(websocket.TextMessage, toSend1)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				var ms viewmodule.WSMsg
				log.Errorf("get token error: %s", err)
				ms.Type = "log"
				ms.Data = fmt.Sprintf("<<\nrequesting error: %s", err)
				toSend, _ := json.Marshal(&ms)
				ws.WriteMessage(websocket.TextMessage, toSend)
				continue
			}
			var ms viewmodule.WSMsg
			ms.Type = "log"
			ms.Data = "<<\n" + resp.Proto + " " + resp.Status + "\n"
			for k, v := range resp.Header {
				ms.Data += k + ": " + strings.Join(v, ",") + "\n"
			}
			ms.Data += "\n"

			body, _ := ioutil.ReadAll(resp.Body)
			ms.Data += string(body)
			toSend, _ := json.Marshal(&ms)
			ws.WriteMessage(websocket.TextMessage, toSend)

			resp.Body.Close()
			mm := make(map[string]interface{})
			json.Unmarshal(body, &mm)
			mm["type"] = "token"
			toSend, _ = json.Marshal(&mm)
			ws.WriteMessage(websocket.TextMessage, toSend)

		case <-time.After(time.Second * 10):
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
