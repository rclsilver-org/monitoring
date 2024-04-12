package server

import (
	"net"
	"net/http"
)

func getRemoteAddress(r *http.Request) string {
	if forwarded_for, ok := r.Header["X-Forwarded-For"]; ok {
		return forwarded_for[0]
	}
	remote_address, _, _ := net.SplitHostPort(r.RemoteAddr)
	return remote_address
}
