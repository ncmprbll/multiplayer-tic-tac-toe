package game

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ncmprbll/multiplayer-tic-tac-toe/types"
)

const (
	FIELD_NOT_SET uint8 = iota
	FIELD_X
	FIELD_O
)

const (
	GAME_NOT_STARTED uint8 = iota
	GAME_WAITING_FOR_X
	GAME_WAITING_FOR_O
	GAME_OVER
)

// 0 - not set
// 1 - x
// 2 - o
type Field struct {
	state uint8
}

func (f Field) Set(value uint8) {
	f.state = value
}

type Grid [3][3]Field

type Game struct {
	Id uuid.UUID
	X  uuid.UUID
	O  uuid.UUID

	State uint8
	Grid  Grid

	Conns []*websocket.Conn
}

var Games = make(map[string]*Game)

func (g Game) Place(x, y int, value uint8) error {
	if x < 0 || x >= len(g.Grid) {
		return errors.New("invalid x value for a grid")
	}

	if y < 0 || y >= len(g.Grid[x]) {
		return errors.New("invalid y value for a grid")
	}

	g.Grid[x][y].Set(value)

	message := types.Message {
		"action": "move",
		"x": x,
		"y": y,
		"value": value,
	}

	g.Broadcast(message)

	return nil
}

func (g Game) IsState(state uint8) bool {
	return g.State == state
}

func (g Game) Broadcast(msg types.Message) {
	for _, c := range g.Conns {
		err := c.WriteJSON(msg)

		if err != nil {
			continue
		}
	}
}

func NewGame() Game {
	game := Game{
		Id: uuid.New(),
	}

	return game
}
