package players

import (
	"context"
	"time"

	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	Id                 string            `bson:"_id,omitempty"`
	Name               string            `bson:"name"`
	Discriminator      string            `bson:"discriminator"`
	SshPubKeys         map[string]string `bson:"ssh_pub_keys"`
	LanguagePreference string            `bson:"language_preference"`
	LastConnectedAt    *time.Time        `bson:"last_connected_at,omitempty"`
	Visitor            bool              `bson:"visitor"`

	color     lipgloss.Color
	turnOrder int
	connected bool
	score     int `bson:"score"`
	isHost    bool
	board     []dice.DicePool
	pool      dice.DicePool

	sess         ssh.Session
	updateChan   chan struct{}
	onDisconnect []func()
	ctx          context.Context
}

func (p *Player) Save() error {
	if len(p.SshPubKeys) == 0 {
		return nil
	}

	opts := options.Replace().SetUpsert(true)
	_, err := database.GetDB().Collection(database.CollectionPlayers).ReplaceOne(p.ctx, bson.M{"_id": p.Id}, p, opts)
	return err
}

func (p *Player) SetName(name string) *Player {
	p.Name = name
	return p
}

func (p *Player) StyledPlayerName(style lipgloss.Style) string {
	if p == nil {
		return ""
	}
	return style.Foreground(p.color).Render(p.Name)
}

func (p *Player) SetTurnOrder(order int) *Player {
	p.turnOrder = order
	return p
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) MakeHost() {
	p.isHost = true
}

func (p *Player) IsHost() bool {
	return p.isHost
}
