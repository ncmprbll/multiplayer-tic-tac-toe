package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
