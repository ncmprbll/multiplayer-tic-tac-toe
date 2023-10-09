package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ncmprbll/multiplayer-tic-tac-toe/api"
)

var APPLICATION_PORT = "1337"

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

	if len(os.Args) > 1 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			APPLICATION_PORT = os.Args[1]
		} else {
			log.Printf("Invalid port as an argument, reverting to default port %v\n", APPLICATION_PORT)
		}
	}

	log.Printf("Listening on port %v\n", APPLICATION_PORT)
	http.ListenAndServe(":"+APPLICATION_PORT, r)
}
