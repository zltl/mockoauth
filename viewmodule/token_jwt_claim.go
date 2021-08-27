package viewmodule

import "github.com/dgrijalva/jwt-go"

type TokenJWTClaim struct {
	jwt.StandardClaims
	Username  string `json:"username"`
	ClientID  string `json:"client_id"`
	ID        string `json:"id"`
	IsRefresh bool   `json:"is_refresh"`
	Scope     string `json:"scope"`
}
