package server

import (
	"net"
	"net/http"
)

//go:generate counterfeiter . Conn
type Conn interface {
	net.Conn
}

//go:generate counterfeiter . HijackerResponseWriter
type HijackerResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
}
