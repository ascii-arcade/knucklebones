package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection string

const (
	CollectionPlayers Collection = "players"
	CollectionGames   Collection = "games"
	CollectionActions Collection = "actions"
)

type MongoDB struct {
	uri      string
	database string

	collections map[string]*mongo.Collection

	client *mongo.Client
	ctx    context.Context
}

var db *MongoDB

func Setup(ctx context.Context, uri, database string) error {
	db = &MongoDB{
		uri:         uri,
		database:    database,
		ctx:         ctx,
		client:      &mongo.Client{},
		collections: map[string]*mongo.Collection{},
	}

	var err error
	db.client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	_ = db.client.Database(db.database).CreateCollection(ctx, string(CollectionPlayers))
	db.collections[string(CollectionPlayers)] = db.client.Database(db.database).Collection(string(CollectionPlayers))

	_ = db.client.Database(db.database).CreateCollection(ctx, string(CollectionGames))
	db.collections[string(CollectionGames)] = db.client.Database(db.database).Collection(string(CollectionGames))

	opts := options.CreateCollection().SetTimeSeriesOptions(options.TimeSeries().SetTimeField("time"))
	_ = db.client.Database(db.database).CreateCollection(ctx, string(CollectionActions), opts)
	db.collections[string(CollectionActions)] = db.client.Database(db.database).Collection(string(CollectionActions))

	if err := createIndexes(ctx); err != nil {
		return err
	}

	return nil
}

func createIndexes(ctx context.Context) error {
	playersCollection := db.collections[string(CollectionPlayers)]

	_, err := playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]any{
			"ssh_pub_keys.$**": 1,
		},
		Options: options.Index().SetName("ssh_pub_keys_wildcard"),
	})

	return err
}

func GetDB() *MongoDB {
	return db
}

func (m *MongoDB) Collection(name Collection) *mongo.Collection {
	return m.collections[string(name)]
}

func (m *MongoDB) Close() error {
	return m.client.Disconnect(m.ctx)
}

func (m *MongoDB) Context() context.Context {
	return m.ctx
}
