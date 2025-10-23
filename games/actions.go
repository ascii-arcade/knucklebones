package games

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/ascii-arcade/knucklebones/players"
)

type ActionType string

const (
	ActionTypeStart ActionType = "start"
	ActionTypeRoll  ActionType = "roll"
	ActionTypePlace ActionType = "place"
)

type Action struct {
	GameID     string          `bson:"game_id"`
	PlayerID   string          `bson:"player_id"`
	Action     ActionType      `bson:"action"`
	DicePool   dice.DicePool   `bson:"dice_pool,omitempty"`
	DiceLocked []dice.DicePool `bson:"dice_locked,omitempty"`
	Turn       int             `bson:"turn"`
	Timestamp  time.Time       `bson:"time"`
}

func (a *Action) Save() error {
	collection := database.GetDB().Collection(database.CollectionActions)
	_, err := collection.InsertOne(database.GetDB().Context(), a)
	return err
}

func (g *Game) saveAction(actionType ActionType) {
	action := Action{
		GameID:    g.Id,
		PlayerID:  g.getTurnPlayer().Id,
		Action:    actionType,
		Turn:      g.Turn,
		Timestamp: time.Now().UTC(),
	}
	if err := action.Save(); err != nil {
		slog.Error("error saving action", "error", err)
	}

	if err := g.Save(); err != nil {
		slog.Error("error saving game after action", "error", err)
	}
}

func (g *Game) Begin() error {
	return g.withErrLock(func() error {
		if error := g.IsPlayerCountOk(); error != nil {
			return error
		}

		g.InProgress = true
		g.saveAction(ActionTypeStart)
		return nil
	})
}

func (g *Game) Reset() {
	g.withLock(func() {
		g.playerOneData().ResetBoard()
		g.playerOneData().ResetPool()

		g.playerTwoData().ResetBoard()
		g.playerTwoData().ResetPool()

		g.rolled = false
		g.Finished = false
		g.Turn = 0
	})
}

func (g *Game) RollDice(rolling bool) {
	g.withLock(func() {
		switch g.Turn {
		case 0:
			g.playerOneData().Pool().Roll()
		case 1:
			g.playerTwoData().Pool().Roll()
		}

		if !rolling {
			g.rolled = true
		}

		g.saveAction(ActionTypeRoll)
	})
}

func (g *Game) PlaceDie(p *players.Player, column int) error {
	return g.withErrLock(func() error {
		if !g.IsTurn(p) {
			return ErrNotYourTurn
		}

		if !g.rolled {
			return ErrDiceNotRolled
		}

		data := g.GetPlayerData(p)
		spot := nextSpot(data.board[column])
		if spot == -1 {
			return ErrColumnFull
		}

		data.board[column][spot] = data.pool[0]
		data.pool = make(dice.DicePool, 1)

		oData := g.GetPlayerData(g.GetOpponent(p))
		removeSame(oData.board[column], data.board[column][spot])

		if full(data.board) {
			g.Finished = true
			return nil
		}

		g.nextTurn()
		g.saveAction(ActionTypePlace)
		return nil
	})
}
