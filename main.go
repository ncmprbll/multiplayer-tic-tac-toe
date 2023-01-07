package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ncmprbll/multiplayer-tic-tac-toe/api"
)

const APPLICATION_PORT = "1339"

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)

	mimes := http.FileServer(http.Dir("web/"))

	r.Handle("/css/*", mimes)
	r.Handle("/js/*", mimes)

	r.Get("/", api.RootHandler)

	r.Post("/play", api.CreateGameHandler)
	r.Get("/play", api.GetGameHandler)
	r.Get("/play/{id:*}", api.GetGameHandler)

	r.Get("/ws", api.WsHandler)
	r.Get("/ws/{id:*}", api.WsHandler)

	http.ListenAndServe(":"+APPLICATION_PORT, r)
}
