package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// 0 - not set
// 1 - x
// 2 - o
type Field struct {
	state int8
}

type Grid [3][3]Field

type Game struct {
	id uuid.UUID
	x  uuid.UUID
	o  uuid.UUID
	conns []*websocket.Conn
}

func (g Game) Broadcast(msg map[string]interface{}) {
	for _, c := range g.conns {
		c.WriteJSON(msg)
	}
}

var Games = make(map[string]*Game)

func NewGame() Game {
	game := Game{
		id: uuid.New(),
	}

	return game
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/")
}

func createGameHandler(w http.ResponseWriter, r *http.Request) {
	game := NewGame()
	id := game.id.String()
	Games[id] = &game

	http.Redirect(w, r, "play/"+id, http.StatusSeeOther)
}

func getGameHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	game, ok := Games[id]

	if !ok {
		http.Error(w, "Bad game ID", http.StatusNotFound)
		return
	}

	pid := ""
	cookie, err := r.Cookie("player-id")

	if err == nil {
		pid = cookie.Value
	}

	fmt.Println("test", pid)

	if (game.x == uuid.Nil || game.o == uuid.Nil) && pid != game.x.String() && pid != game.o.String() {
		id := uuid.New()

		if game.x == uuid.Nil {
			game.x = id
		} else if game.o == uuid.Nil {
			game.o = id
		}

		cookie := &http.Cookie{
			Name:  "player-id",
			Value: id.String(),
		}

		http.SetCookie(w, cookie)
	}

	// fmt.Fprintf(w, "Game: %v %v", game.x, game.o)
	http.ServeFile(w, r, "public/game.html")
}

func processAction(msg map[string]interface{}, game *Game) {
	if msg["action"] == "Move" {
		text := "X"

		if msg["player"] == game.o.String() {
			text = "O"
		}

		msg["text"] = text
		game.Broadcast(msg)
	}
}

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer ws.Close()

	id := chi.URLParam(r, "id")

	game, ok := Games[id]

	if !ok {
		ws.WriteMessage(1, []byte("Bad game ID: " + id))
		return
	}

	game.conns = append(game.conns, ws)

	for {
		var msg map[string]interface{}

		err := ws.ReadJSON(&msg)

		if err != nil {
			fmt.Println(err)
			continue
		}

		processAction(msg, game)
		fmt.Println(ws.LocalAddr(), msg["action"])//string(p))
	}
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)

	r.Get("/", rootHandler)
	r.Post("/play", createGameHandler)
	r.Get("/play", getGameHandler)
	r.Get("/play/{id:*}", getGameHandler)
	r.Get("/ws", wsHandler)
	r.Get("/ws/{id:*}", wsHandler)

	http.ListenAndServe(":1337", r)
}
