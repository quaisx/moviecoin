package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://root:password@localhost:27017/?maxPoolSize=20&w=majority"

// Starting point - https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection/
// Docker running mongodb - please map the port and reflect that in the connection uri
// To create a user in mongodb instance running in a docker container
// use admin
// db.createUser({user:"root",pwd:"xxxx",roles:["root"]})
type Client struct {
	mongo *mongo.Client
	ctx   context.Context
}

func (db *Client) Disconnect() {
	if db != nil && db.mongo != nil {
		if err := db.mongo.Disconnect(db.ctx); err != nil {
			log.Fatal(err)
		}
	}
}

func NewClient(uri string) *Client {
	client := new(Client)
	x, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	client.mongo = x
	client.ctx, _ = context.WithTimeout(context.Background(), 300*time.Second)

	err = client.mongo.Connect(client.ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (db *Client) InsertX(collection *mongo.Collection, docs []interface{}) *mongo.InsertManyResult {
	res, insertErr := collection.InsertMany(db.ctx, docs)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	return res
}

func (db *Client) Insert(collection *mongo.Collection, doc []interface{}) *mongo.InsertOneResult {
	res, insertErr := collection.InsertOne(db.ctx, doc)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	return res
}

func (db *Client) Find(collection *mongo.Collection, filter interface{}, shelf interface{}) {
	if err := collection.FindOne(db.ctx, filter).Decode(shelf); err != nil {
		log.Fatal(err)
	}
}

// @TODO: Refactor to accept Go struct to match the results
func (db *Client) FindX(collection *mongo.Collection, filter interface{}, shelves interface{}) {
	cur, err := collection.Find(db.ctx, filter)
	defer cur.Close(db.ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err = cur.All(db.ctx, shelves); err != nil {
		panic(err)
	}
}
