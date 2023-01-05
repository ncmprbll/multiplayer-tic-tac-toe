package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"errors"

	"github.com/ncmprbll/multiplayer-tic-tac-toe/game"
	"github.com/ncmprbll/multiplayer-tic-tac-toe/types"
)

func processAction(msg types.Message, g *game.Game) error {
	action, ok := msg["action"]

	if !ok {
		return errors.New("no action to process")
	}

	player, ok := msg["player"]

	if !ok {
		return errors.New("no player")
	}

	if action == game.ACTION_MOVE {
		x, ok := msg["x"]

		if !ok {
			return errors.New("no x value")
		}

		y, ok := msg["y"]

		if !ok {
			return errors.New("no y value")
		}

		var value uint8

		if player == g.X.String() {
			value = game.FIELD_X
		} else if player == g.O.String() {
			value = game.FIELD_O
		} else {
			return errors.New("non-player making a move")
		}

		xVal, ok := x.(float64)

		if !ok {
			return errors.New("x not a number")
		}

		yVal, ok := y.(float64)

		if !ok {
			return errors.New("y not a number")
		}

		xInt, yInt := int(xVal), int(yVal)

		g.Place(xInt, yInt, value)
	} else if action == game.ACTION_CHAT {
		text, ok := msg["text"]

		if !ok {
			return errors.New("no text value")
		}

		g.SendChatMessage(player.(string), text.(string))
	}

	return nil
}

var upgrader = websocket.Upgrader{}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	g, ok := game.Games[id]

	if !ok || g.IsState(game.GAME_OVER) {
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	player := r.URL.Query().Get("id")

	if player == g.X.String() && !g.XAlive {
		g.XAlive = true
		g.SendSystemMessage("X has reconnected")
	} else if player == g.O.String() && !g.OAlive {
		g.OAlive = true
		g.SendSystemMessage("O has reconnected")
	}

	ws.SetCloseHandler(func(code int, text string) error {
		if g.IsState(game.GAME_OVER) {
			return nil
		}

		if g.IsState(game.GAME_NOT_STARTED) {
			g.X = uuid.Nil
			g.XAlive = false
			g.O = uuid.Nil
			g.OAlive = false

			return nil
		}

		if player == g.X.String() {
			g.XAlive = false
			g.SendSystemMessage("X has disconnected")
		} else if player == g.O.String() {
			g.OAlive = false
			g.SendSystemMessage("O has disconnected")
		}

		if !g.XAlive && !g.OAlive {
			g.SendSystemMessage("No players left, deleting the game...")
			g.Over()
		}

		return nil
	})

	g.Conns = append(g.Conns, ws)

	g.ConnectionUpdate(ws)
	g.BroadcastState()

	defer ws.Close()

	readErrCount := 0

	for {
		var msg types.Message

		err := ws.ReadJSON(&msg)

		if readErrCount > 512 {
			return
		}

		if err != nil {
			readErrCount++
			continue
		}

		processAction(msg, g)
	}
}
