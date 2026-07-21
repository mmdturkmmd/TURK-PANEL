// Package global provides global variables and interfaces for accessing web server.
package global

import (
	"context"
	"net/http"
	_ "unsafe"

	"github.com/robfig/cron/v3"
)

var webServer WebServer

// WebServer interface defines methods for accessing the web server instance.
// NOTE: Do NOT add methods returning types from packages that import web.*;
// that creates cycle: web → sub → web/service → web/global → sub
type WebServer interface {
	GetCron() *cron.Cron
	GetCtx() context.Context
	GetWSHub() any
	GetHttpServer() *http.Server
	RestartXray() error
}

// SetWebServer sets the global web server instance.
func SetWebServer(s WebServer) { webServer = s }

// GetWebServer returns the global web server instance.
func GetWebServer() WebServer { return webServer }
