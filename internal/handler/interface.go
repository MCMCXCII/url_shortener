package handler

import "net/http"

type Handler interface {
	Routes() http.Handler
}
