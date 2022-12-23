package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ncmprbll/multiplayer-tic-tac-toe/game"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/")
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	g := game.NewGame()
	id := g.Id.String()
	game.Games[id] = &g

	http.Redirect(w, r, "play/"+id, http.StatusSeeOther)
}

func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	g, ok := game.Games[id]

	if !ok {
		http.Error(w, "Bad game ID", http.StatusNotFound)
		return
	}

	pid := ""
	cookie, err := r.Cookie("player-id")

	if err == nil {
		pid = cookie.Value
	}

	if (g.X == uuid.Nil || g.O == uuid.Nil) && pid != g.X.String() && pid != g.O.String() {
		id := uuid.New()

		if g.X == uuid.Nil {
			g.X = id
		} else if g.O == uuid.Nil {
			g.O = id
		}

		cookie := &http.Cookie{
			Name:  "player-id",
			Value: id.String(),
		}

		http.SetCookie(w, cookie)
	}

	http.ServeFile(w, r, "public/game.html")
}