package controller

import (
	"fmt"
	"mockoauth/viewmodule"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Token godoc
// @Summary oauth2 token request
// @Description request token by code
// @Produce  html
// @Router /oauth2/token/:ID [get]
func (s *Controller) Token(c *gin.Context) {

	c.Request.ParseForm()
	ID := c.Param("ID")
	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")

	s.mu.Lock()
	wsConn, ok := s.wsMap[ID]
	s.mu.Unlock()
	if ok {
		logs := fmt.Sprintf("POST %s\n\n", c.Request.RequestURI)
		for k, v := range c.Request.Form {
			logs += fmt.Sprintf("%s %s\n", k, v[0])
		}
		wsConn.ch <- logs
	}

	if grantType == "authorization_code" {
		claimi, err := jwt.ParseWithClaims(code, &viewmodule.CodeJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		})
		if err != nil {
			log.Error(err)
			c.String(http.StatusUnauthorized, "")
			return
		}
		codeClaim := claimi.Claims.(*viewmodule.CodeJWTClaim)
		log.Infof("getting token ID=%s, username=%s, clientid=%s",
			ID, codeClaim.Username, codeClaim.ClientID)
		if redirectURI != codeClaim.RedirectURI {
			logs := fmt.Sprintf("redirect uri not match, get %s, expected %s", redirectURI, codeClaim.RedirectURI)
			log.Error(logs)
			if ok {
				wsConn.ch <- logs
			}
			c.String(http.StatusBadRequest, logs)
			return
		}
		if time.Now().Unix() > codeClaim.ExpiresAt {
			logs := fmt.Sprintf("token expired at %d", codeClaim.ExpiresAt)
			log.Error(logs)
			if ok {
				wsConn.ch <- logs
			}
			c.String(http.StatusBadRequest, logs)
			return
		}

		tclaims := &viewmodule.TokenJWTClaim{}
		tokenSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, tclaims)
		tclaims.Username = codeClaim.Username
		tclaims.ClientID = codeClaim.ClientID
		tclaims.Scope = c.PostForm("scope")
		tclaims.ExpiresAt = time.Now().Add(time.Second * 3600).Unix()
		tclaims.IsRefresh = false
		actoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		tclaims.IsRefresh = true
		reftoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		if ok {
			logs := fmt.Sprintf("access_token: %s\nrefresh_token: %s", actoken, reftoken)
			wsConn.ch <- logs
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  actoken,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": reftoken,
		})
	} else if grantType == "password" {
		username := c.PostForm("username")
		password := c.PostForm("password")

		tclaims := &viewmodule.TokenJWTClaim{}
		tokenSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, tclaims)
		tclaims.Username = username
		tclaims.ClientID = "Password-Credentials-Grant"
		tclaims.Scope = c.PostForm("scope")
		tclaims.ExpiresAt = time.Now().Add(time.Second * 3600).Unix()
		tclaims.IsRefresh = false
		actoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		tclaims.IsRefresh = true
		reftoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		if ok {
			logs := fmt.Sprintf("username: %s\npassword:%s\naccess_token: %s\nrefresh_token: %s",
				username, password, actoken, reftoken)
			wsConn.ch <- logs
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  actoken,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": reftoken,
		})
	} else if grantType == "client_credentials" {
		tclaims := &viewmodule.TokenJWTClaim{}
		tokenSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, tclaims)
		tclaims.Username = "none"
		tclaims.ClientID = "Client-Credentials-Grant"
		tclaims.Scope = c.PostForm("scope")
		tclaims.ExpiresAt = time.Now().Add(time.Second * 3600).Unix()
		tclaims.IsRefresh = false
		actoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		tclaims.IsRefresh = true
		reftoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		if ok {
			logs := fmt.Sprintf("access_token: %s\nrefresh_token: %s", actoken, reftoken)
			wsConn.ch <- logs
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  actoken,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": reftoken,
		})
	} else if grantType == "refresh_token" {
		refreshToken := c.PostForm("refresh_token")
		refclaimi, err := jwt.ParseWithClaims(refreshToken, &viewmodule.CodeJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		})
		if err != nil {
			log.Error(err)
			c.String(http.StatusUnauthorized, "")
			return
		}
		refclaim := refclaimi.Claims.(*viewmodule.CodeJWTClaim)

		tclaims := &viewmodule.TokenJWTClaim{}
		tokenSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, tclaims)
		tclaims.Username = refclaim.Username
		tclaims.ClientID = refclaim.ClientID
		tclaims.ID = refclaim.ID
		tclaims.Scope = refclaim.Scope
		tclaims.ExpiresAt = time.Now().Add(time.Second * 3600).Unix()
		tclaims.IsRefresh = false
		actoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))

		tclaims.IsRefresh = true
		reftoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))
		c.JSON(http.StatusOK, gin.H{
			"access_token":  actoken,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": reftoken,
		})

		if ok {
			logs := fmt.Sprintf("username: %s\naccess_token: %s\nrefresh_token: %s",
				tclaims.Username, actoken, reftoken)
			wsConn.ch <- logs
		}
	} else {
		log.Error("unknown grant type")
		c.String(http.StatusBadRequest, "")
	}
}
