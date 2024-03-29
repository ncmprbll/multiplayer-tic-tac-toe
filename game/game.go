package game

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ncmprbll/multiplayer-tic-tac-toe/types"
)

const (
	ACTION_MOVE         = "move"
	ACTION_UPDATE       = "update"
	ACTION_STATE_UPDATE = "state_update"
	ACTION_CHAT         = "chat"
	ACTION_SWITCH       = "switch"
	ACTION_ROUND_END    = "round_end"
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
	GAME_ROUND_END
	GAME_OVER
)

type Grid [3][3]uint8

type Game struct {
	Id uuid.UUID
	X  uuid.UUID
	O  uuid.UUID

	XAlive bool
	OAlive bool

	State uint8
	Grid  Grid

	Conns []*websocket.Conn
	ConnsLock sync.Mutex

	ChatLog []types.Message
}

var (
	Games = make(map[string]*Game)
	Locks = make(map[string]sync.Mutex)
	GamesWLock = &sync.Mutex{}
)

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

	if g.IsState(GAME_NOT_STARTED) || g.IsState(GAME_ROUND_END) || g.IsState(GAME_OVER) {
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
		"action": ACTION_MOVE,
		"x":      x,
		"y":      y,
		"value":  value,
	}

	g.Broadcast(message)

	finished, pattern := g.isFinishingMove(x, y)

	if finished {
		g.State = GAME_ROUND_END

		message := types.Message{
			"action": ACTION_ROUND_END,
			"value":  pattern,
		}

		g.Broadcast(message)
		g.SendSystemMessage("The match is over, switching sides...")

		timer := time.NewTimer(3 * time.Second)

		go func() {
			<-timer.C
			g.SwitchSides()
		}()
	}

	g.BroadcastState()

	return nil
}

func (g *Game) isFinishingMove(x, y int) (bool, [][]int) {
	value := g.Grid[x][y]

	row := g.Grid[x]
	rowLen := len(row)

	// Row victory
	for c, v := range row {
		if v != value {
			break
		} else if c == rowLen-1 && v == value {
			return true, [][]int{{x, 0}, {x, 1}, {x, 2}}
		}
	}

	colLen := len(g.Grid)

	// Column victory
	for c, r := range g.Grid {
		if r[y] != value {
			break
		} else if c == colLen-1 && r[y] == value {
			return true, [][]int{{0, y}, {1, y}, {2, y}}
		}
	}

	// Diagonal victory
	if g.Grid[1][1] == value {
		if g.Grid[0][0] == value && g.Grid[2][2] == value {
			return true, [][]int{{0, 0}, {1, 1}, {2, 2}}
		} else if g.Grid[0][2] == value && g.Grid[2][0] == value {
			return true, [][]int{{0, 2}, {1, 1}, {2, 0}}
		}
	}

	for _, r := range g.Grid {
		for _, v := range r {
			if v == FIELD_NOT_SET {
				return false, nil
			}
		}
	}

	// Round draw
	return true, nil
}

func (g *Game) IsState(state uint8) bool {
	return g.State == state
}

func (g *Game) ConnectionUpdate(c *websocket.Conn) error {
	message := types.Message{
		"action": ACTION_UPDATE,
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
		c.WriteJSON(msg)
	}
}

func (g *Game) SwitchSides() {
	g.Grid = *new(Grid)

	gridupdate := types.Message{
		"action": ACTION_UPDATE,
		"value":  g.Grid,
	}

	switchsides := types.Message{
		"action": ACTION_SWITCH,
	}

	t := g.X

	g.X = g.O
	g.O = t

	g.State = GAME_WAITING_FOR_X

	g.Broadcast(gridupdate)
	g.Broadcast(switchsides)
	g.BroadcastState()
}

func (g *Game) BroadcastState() {
	message := types.Message{
		"action": ACTION_STATE_UPDATE,
		"value":  g.State,
	}

	g.Broadcast(message)
}

func (g *Game) Over() {
	g.State = GAME_OVER
	g.BroadcastState()

	for _, c := range g.Conns {
		c.Close()
	}

	GamesWLock.Lock()
	delete(Games, g.Id.String())
	GamesWLock.Unlock()
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
	} else if !issystem {
		return
	}

	text := types.Message{
		"action":    ACTION_CHAT,
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
