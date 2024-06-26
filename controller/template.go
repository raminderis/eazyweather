package controller

import "net/http"

type Template interface {
	Execute(w http.ResponseWriter, data interface{})
}
