package players

import (
	"context"
	"time"

	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/ascii-arcade/knucklebones/language"
	"github.com/ascii-arcade/knucklebones/utils"
	"github.com/google/uuid"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, pkn, pk, langPref string) (*Player, error) {
	board := make([]dice.DicePool, 3)
	for i := range board {
		board[i] = make(dice.DicePool, 3)
	}

	player := &Player{
		Id:                 uuid.New().String(),
		Name:               utils.GenerateName(language.Languages[langPref]),
		Discriminator:      utils.GenerateDescriminator(),
		SshPubKeys:         map[string]string{pkn: pk},
		LanguagePreference: langPref,

		score:      0,
		updateChan: make(chan struct{}),
		board:      board,
		pool:       make(dice.DicePool, 1),
		ctx:        ctx,
	}

	return player, player.Save()
}

func (p *Player) WithContext(ctx context.Context) *Player {
	p.ctx = ctx
	return p
}

func (p *Player) Connect() {
	p.updateChan = make(chan struct{})
	players[p.Id] = p
	p.OnDisconnect(func() {
		RemovePlayer(p)
	})

	activityTicker := time.NewTicker(5 * time.Second)
	timeoutDuration := config.GetPlayerTimeoutDuration()
	lastActivity := time.Now()

	timeoutCtx, timeoutCancel := context.WithCancel(p.ctx)

	go func() {
		defer activityTicker.Stop()
		defer timeoutCancel()

		for {
			select {
			case <-timeoutCtx.Done():
				return
			case <-activityTicker.C:
				now := time.Now()

				if now.Sub(lastActivity) > timeoutDuration {
					if p.sess != nil {
						p.sess.Close()
					}
					return
				}

				p.LastConnectedAt = utils.ToPointer(now)
				_ = p.Save()
			case <-p.updateChan:
				lastActivity = time.Now()
			}
		}
	}()

	go func() {
		<-timeoutCtx.Done()
		for _, fn := range p.onDisconnect {
			fn()
		}
	}()
}

func Get(sshPubKey string) (*Player, bool) {
	pipeline := []map[string]any{
		{
			"$match": map[string]any{
				"$expr": map[string]any{
					"$gt": []any{
						map[string]any{
							"$size": map[string]any{
								"$filter": map[string]any{
									"input": map[string]any{"$objectToArray": "$ssh_pub_keys"},
									"cond":  map[string]any{"$eq": []any{"$$this.v", sshPubKey}},
								},
							},
						},
						0,
					},
				},
			},
		},
	}

	cursor, err := database.GetDB().Collection(database.CollectionPlayers).Aggregate(context.Background(), pipeline)
	if err == nil {
		defer cursor.Close(context.Background())
		if cursor.Next(context.Background()) {
			var player Player
			if err := cursor.Decode(&player); err == nil {
				return &player, true
			}
		}
	}

	return nil, false
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.Id]; exists {
		close(player.updateChan)
		delete(players, player.Id)
	}
}

func GetPlayerCount() int {
	return len(players)
}

func GetConnectedPlayerCount() int {
	count := 0
	for _, player := range players {
		if player.connected {
			count++
		}
	}
	return count
}
