package handler

import (
	"net/http"

	"github.com/pajtaand/dmap-zero/webapp"
)

type webAppHandler struct{}

func NewWebAppHandler() *webAppHandler {
	return &webAppHandler{}
}

func (h *webAppHandler) GetWebApplication(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	app := webapp.GetApp()
	if _, err := w.Write(app); err != nil {
		panic(err)
	}
}

func (h *webAppHandler) GetFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")

	app := webapp.GetFavicon()
	if _, err := w.Write(app); err != nil {
		panic(err)
	}
}

func (h *webAppHandler) GetIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")

	app := webapp.GetIcon()
	if _, err := w.Write(app); err != nil {
		panic(err)
	}
}

func (h *webAppHandler) GetCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")

	app := webapp.GetCSS()
	if _, err := w.Write(app); err != nil {
		panic(err)
	}
}

func (h *webAppHandler) GetJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")

	app := webapp.GetJS()
	if _, err := w.Write(app); err != nil {
		panic(err)
	}
}
