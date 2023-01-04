package game

import (
	"errors"
	"strings"
	"time"

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

type Grid [3][3]uint8

type Game struct {
	Id uuid.UUID
	X  uuid.UUID
	O  uuid.UUID

	State uint8
	Grid  Grid

	Conns []*websocket.Conn

	ChatLog []types.Message
}

var Games = make(map[string]*Game)

func (g *Game) Place(x, y int, value uint8) error {
	if x < 0 || x >= len(g.Grid) {
		return errors.New("invalid x value for a grid")
	}

	if y < 0 || y >= len(g.Grid[x]) {
		return errors.New("invalid y value for a grid")
	}

	if g.Grid[x][y] != FIELD_NOT_SET {
		return errors.New("invalid field")
	}

	if g.IsState(GAME_NOT_STARTED) || g.IsState(GAME_OVER) {
		return errors.New("the game has not started or is over")
	}

	if value == FIELD_NOT_SET {
		return errors.New("trying to unset a field")
	}

	isX := g.IsState(GAME_WAITING_FOR_X)
	isO := g.IsState(GAME_WAITING_FOR_O)

	if (isX && value != FIELD_X) || (isO && value != FIELD_O) {
		return errors.New("invalid turn")
	}

	if isX {
		g.State = GAME_WAITING_FOR_O
	}

	if isO {
		g.State = GAME_WAITING_FOR_X
	}

	g.Grid[x][y] = value

	message := types.Message{
		"action": "move",
		"x":      x,
		"y":      y,
		"value":  value,
	}

	g.Broadcast(message)

	finished, pattern := g.isFinishingMove(x, y)

	if finished {
		g.State = GAME_OVER

		message := types.Message{
			"action": "victory",
			"value":  pattern,
		}

		g.Broadcast(message)
		g.SendSystemMessage("The match is over")
	}

	g.BroadcastState()

	return nil
}

func (g *Game) isFinishingMove(x, y int) (bool, [][]int) {
	value := g.Grid[x][y]

	row := g.Grid[x]
	rowLen := len(row)

	for c, v := range row {
		if v != value {
			break
		} else if c == rowLen-1 && v == value {
			return true, [][]int{{x, 0}, {x, 1}, {x, 2}}
		}
	}

	colLen := len(g.Grid)

	for c, r := range g.Grid {
		if r[y] != value {
			break
		} else if c == colLen-1 && r[y] == value {
			return true, [][]int{{0, y}, {1, y}, {2, y}}
		}
	}

	if g.Grid[1][1] != value {
		return false, nil
	}

	if g.Grid[0][0] == value && g.Grid[2][2] == value {
		return true, [][]int{{0, 0}, {1, 1}, {2, 2}}
	} else if g.Grid[0][2] == value && g.Grid[2][0] == value {
		return true, [][]int{{0, 2}, {1, 1}, {2, 0}}
	}

	return false, nil
}

func (g *Game) IsState(state uint8) bool {
	return g.State == state
}

func (g *Game) ConnectionUpdate(c *websocket.Conn) error {
	message := types.Message{
		"action": "update",
		"value":  g.Grid,
	}

	if err := c.WriteJSON(message); err != nil {
		return err
	}

	for _, m := range g.ChatLog {
		if err := c.WriteJSON(m); err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) Broadcast(msg types.Message) {
	for _, c := range g.Conns {
		err := c.WriteJSON(msg)

		// Does it make any sense though? I think I was planning on handling an error here somehow
		// I'll keep it this way for the time being
		if err != nil {
			continue
		}
	}
}

func (g *Game) BroadcastState() {
	message := types.Message{
		"action": "state_update",
		"value":  g.State,
	}

	g.Broadcast(message)
}

func (g *Game) chatMessage(player string, message string, issystem bool) {
	// Removing trailing and repeated spaces in case of a client-side bypass
	formatted := strings.Join(strings.Fields(strings.TrimSpace(message)), " ")

	if formatted == "" {
		return
	}

	// TODO: Come up with a decent name system
	sender := "System"

	if player == g.X.String() {
		sender = "X"
	} else if player == g.O.String() {
		sender = "O"
	}

	text := types.Message{
		"action":    "chat",
		"timestamp": time.Now().Format("15:04:05"),
		"text":      formatted,
		"sender":    sender,
		"issystem":  issystem,
	}

	g.Broadcast(text)
	g.ChatLog = append(g.ChatLog, text)
}

func (g *Game) SendChatMessage(player string, message string) {
	g.chatMessage(player, message, false)
}

func (g *Game) SendSystemMessage(message string) {
	g.chatMessage("", message, true)
}

func NewGame() Game {
	game := Game{
		Id: uuid.New(),
	}

	return game
}
