package controller

import (
	"encoding/base64"
	"mockoauth/viewmodule"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Authorize godoc
// @Summary oauth2 authorize login page
// @Description show logs and details of Authorization Code Grant oauth mode
// @Produce  html
// @Router /oauth2/authorize/:ID [get]
func (s *Controller) Authorize(c *gin.Context) {
	clientInfo := viewmodule.ClientInfo{}
	host := c.Request.Host
	ID := c.Param("ID")
	clientInfo.ID = ID

	grantType := c.Request.URL.Query().Get("grant_type")
	redirectURI := c.Request.URL.Query().Get("redirect_uri")
	urlp, _ := url.Parse(redirectURI)

	log.Infof("ID=%s", ID)
	log.Infof("grant_type=%s, requestURI=%s", grantType, c.Request.RequestURI)

	oauthProviderInfo := viewmodule.OAuthProviderInfo{
		ClientID:     clientInfo.ID,
		ClientSecret: base64.RawURLEncoding.EncodeToString([]byte(clientInfo.ID)),
	}

	var params []viewmodule.Pair
	for k, v := range c.Request.URL.Query() {
		log.Infof("%s: %s", k, v[0])
		params = append(params, viewmodule.Pair{Key: k, Value: v[0]})
	}

	c.HTML(http.StatusOK, "tmpl/authorize.tmpl",
		gin.H{
			"title":             "Authorization Code Grant Mock",
			"oauthProviderInfo": oauthProviderInfo,
			"host":              host,
			"client":            clientInfo,
			"params":            params,
			"redirectHost":      urlp.Host,
		})
}

// Authorize godoc
// @Summary oauth2 authorize login page
// @Description show logs and details of Authorization Code Grant oauth mode
// @Produce  html
// @Router /oauth2/acg/authorize/:ID [post]
func (s *Controller) AuthorizePost(c *gin.Context) {
	c.Request.ParseForm()
	ID := c.Param("ID")
	username := c.PostForm("username")
	password := c.PostForm("password")
	redirectURI := c.PostForm("redirect_uri")
	responseType := c.PostForm("response_type")

	nextURLValues := url.Values{}
	for k, v := range c.Request.PostForm {
		if k == "username" || k == "password" || k == "redirect_uri" {
			continue
		}
		if k == "client_id" || k == "client_secret" || k == "response_type" {
			continue
		}
		nextURLValues.Add(k, v[0])
	}

	if responseType == "code" {
		claims := &viewmodule.CodeJWTClaim{}
		codeSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		claims.Username = username
		claims.ClientID = c.PostForm("client_id")
		claims.RedirectURI = redirectURI
		claims.Scope = c.PostForm("scope")
		claims.ExpiresAt = time.Now().Add(time.Minute * 10).Unix()
		code, _ := codeSrc.SignedString([]byte(s.jwtSecret))

		nextURLValues.Add("code", code)

		nextURL := redirectURI + "?" + nextURLValues.Encode()

		log.Infof("ID=%s", ID)
		log.Infof("username=%s, password=%s, code=%s", username, password, code)

		c.Redirect(http.StatusFound, nextURL)
	} else if responseType == "token" {
		tclaims := &viewmodule.TokenJWTClaim{}
		tokenSrc := jwt.NewWithClaims(jwt.SigningMethodHS256, tclaims)
		tclaims.Username = username
		tclaims.ClientID = ID
		tclaims.Scope = c.PostForm("scope")
		tclaims.ExpiresAt = time.Now().Add(time.Second * 3600).Unix()
		tclaims.IsRefresh = false
		actoken, _ := tokenSrc.SignedString([]byte(s.jwtSecret))
		log.Infof("ID=%s", ID)
		log.Infof("username=%s, password=%s, token=%s", username, password, actoken)

		nextURLValues.Add("access_token", actoken)
		nextURLValues.Add("token_type", "bearer")
		nextURLValues.Add("expires_in", "3600")
		nextURL := redirectURI + "#" + nextURLValues.Encode()
		c.Redirect(http.StatusFound, nextURL)

	}
}
