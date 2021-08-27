package main

import (
	"mockoauth/server"
	"net/http"
	"os"

	_ "net/http/pprof"

	_ "mockoauth/docs"
	_ "mockoauth/resources"

	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/urfave/cli/v2"
)

// @title Mock Oauth2 API
// @version 1.0
// @description This is an oauth mock application.

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information

// @x-extension-openapi {"example": "value on a json format"}

var (
	// Version is the version of upd tool chains, auto generate when building.
	Version = "manual build has no version"
	// GitHash is git commit hash, auto asigned when building.
	GitHash = "manual build has no git hash"
	// BuildTime is time when tool chains been build, auto asigned when building.
	BuildTime = "manual build has no build time"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
}

func main() {
	app := &cli.App{
		Name:   "gooauth",
		Usage:  "mockoauth server",
		Action: start,
		Authors: []*cli.Author{{
			Name:  "liaotonglang",
			Email: "liaotonglang@gmail.com",
		}},
		Version:   Version,
		Copyright: "(c) 2021 liaotonglang",
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Value:   ":11222",
			Usage:   "listen on ip:port",
			EnvVars: []string{"MOCKOAUTH_LISTEN"},
		},
		&cli.StringFlag{
			Name: "host",
			Value: "localhost",
			Usage: "host to access server",
			EnvVars: []string{"MOCKOAUTH_HOST"},
		},
		&cli.StringFlag{
			Name:    "pprof_listen",
			Value:   "",
			Usage:   "pprof listen on",
			EnvVars: []string{"MOCKOAUTH_PPROF_LISTEN"},
		},
	}

	app.Run(os.Args)
}

func start(c *cli.Context) error {
	log.Println("starting")
	if c.String("pprof_listen") != "" {
		go func() {
			http.ListenAndServe(c.String("pprof_listen"), nil)
		}()
	}

	s := server.NewHTTPServer()
	s.ListenOn = c.String("listen")

	s.Router().GET("/swagger/*any",
		ginSwagger.DisablingWrapHandler(swaggerFiles.Handler,
			"NAME_OF_ENV_VARIABLE"))

	s.Start()

	return nil
}
