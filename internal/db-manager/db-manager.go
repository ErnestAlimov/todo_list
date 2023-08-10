package dbmanager

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DbManager struct {
	TaskCollection *mongo.Collection
}

func NewDbManager(
	dbHost string,
	dbPort string,
	dbName string,
	collectionName string,
) *DbManager {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(dbName)
	err = db.CreateCollection(context.Background(), collectionName)
	if err != nil {
		log.Print(err)
	}

	return &DbManager{
		TaskCollection: db.Collection(collectionName),
	}
}
