package api

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ncmprbll/multiplayer-tic-tac-toe/game"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/")
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	g := game.NewGame()
	id := g.Id.String()

	game.GamesWLock.Lock()
	game.Games[id] = &g
	game.Locks[id] = sync.Mutex{}
	game.GamesWLock.Unlock()

	http.Redirect(w, r, "play/"+id, http.StatusSeeOther)
}

func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	gameid := chi.URLParam(r, "id")

	g, ok := game.Games[gameid]

	if !ok {
		http.Error(w, "Bad game ID", http.StatusNotFound)
		return
	}

	lock := game.Locks[gameid]
	pid := ""
	cookie, err := r.Cookie(gameid + "_id")

	if err == nil {
		pid = cookie.Value
	}

	lock.Lock()
	if (g.X == uuid.Nil || g.O == uuid.Nil) && pid != g.X.String() && pid != g.O.String() {
		id := uuid.New()

		whoami := ""

		if g.X == uuid.Nil {
			g.X = id
			g.XAlive = true

			whoami = "X"
		} else if g.O == uuid.Nil {
			g.O = id
			g.OAlive = true

			whoami = "O"
		}

		reflectionCookie := &http.Cookie{
			Name:   gameid + "_whoami",
			Value:  whoami,
			MaxAge: 3600,
		}

		http.SetCookie(w, reflectionCookie)

		cookie := &http.Cookie{
			Name:   gameid + "_id",
			Value:  id.String(),
			MaxAge: 3600,
		}

		http.SetCookie(w, cookie)
	}

	if g.State == game.GAME_NOT_STARTED && g.X != uuid.Nil && g.O != uuid.Nil {
		g.State = game.GAME_WAITING_FOR_X
		g.SendSystemMessage("The game has begun")
	}
	lock.Unlock()

	http.ServeFile(w, r, "web/game.html")
}
