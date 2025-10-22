package players

import (
	"context"
	"time"

	"github.com/ascii-arcade/knucklebones/database"
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

func (p *Player) IsConnected() bool {
	_, ok := players[p.Id]
	return ok
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}

func (p *Player) UpdateChan() chan struct{} {
	if p.updateChan == nil {
		p.updateChan = make(chan struct{}, 1)
	}
	return p.updateChan
}

func (p *Player) SignalActivity() {
	if p.updateChan != nil {
		select {
		case p.updateChan <- struct{}{}:
		default:
		}
	}
}
