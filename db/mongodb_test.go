package db

import (
	"fmt"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConnection(t *testing.T) {
	client := NewClient(uri)
	defer client.Disconnect()

	// Interact with data
	type Entry struct {
		ID      primitive.ObjectID `bson:"_id,omitempty"`
		Title   string             `bson:"title,omitempty"`
		Content string             `bson:"content,omitempty"`
	}

	collection := client.mongo.Database("test").Collection("entries")

	docs := []interface{}{
		bson.D{{Key: "title", Value: "Title 1"}, {Key: "content", Value: "Document content 1"}},
		bson.D{{Key: "title", Value: "Title 2"}, {Key: "content", Value: "Document content 2"}},
		bson.D{{Key: "title", Value: "Title 3"}, {Key: "content", Value: "Document content 3"}},
	}

	res := client.InsertX(collection, docs)
	log.Println(res)
	/*
		Iterate a cursor
	*/
	var entries []Entry
	client.FindX(collection, bson.M{}, &entries)

	for _, entry := range entries {
		log.Println(entry)
	}
	var oneEntry Entry
	client.Find(collection, bson.M{"title": "Title 1"}, &oneEntry)
	log.Println(oneEntry)

	fmt.Println("done")
}
