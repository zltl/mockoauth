package viewmodule

import "github.com/dgrijalva/jwt-go"

type CodeJWTClaim struct {
	jwt.StandardClaims
	Username    string `json:"username"`
	ClientID    string `json:"client_id"`
	ID          string `json:"id"`
	RedirectURI string `json:"redirect_uri"`
	Scope       string `json:"scope"`
}
