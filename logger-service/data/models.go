package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdateAt  time.Time `bson:"updated_at" json:"updated_at"`
}

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		LogEntry: LogEntry{},
	}
}

func (l *LogEntry) Insert(entry LogEntry) (err error) {
	_, err = client.
		Database("logs").
		Collection("logs").
		InsertOne(nil, LogEntry{
			Name:      entry.Name,
			Data:      entry.Data,
			CreatedAt: time.Now(),
			UpdateAt:  time.Now(),
		})

	return err
}

func (l *LogEntry) All() (entries []*LogEntry, err error) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
		opts        = options.Find().SetSort(bson.D{{"created_at", -1}})
		cursor      *mongo.Cursor
	)
	defer cancel()

	if cursor, err = client.Database("logs").Collection("logs").Find(context.TODO(), bson.D{}, opts); err != nil {
		log.Println("Error finding logs", err)
		return nil, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item LogEntry
		if err = cursor.Decode(&item); err != nil {
			log.Println("Error decoding log", err)
			return nil, err
		}

		entries = append(entries, &item)
	}

	return entries, nil
}
