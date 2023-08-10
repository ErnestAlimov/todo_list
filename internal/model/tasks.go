package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	Active string = "active"
	Done   string = "done"
)

type Task struct {
	ID       primitive.ObjectID `json:"-" bson:"_id"`
	Title    string             `json:"title" bson:"title"`
	ActiveAt string             `json:"activeAt"  bson:"activeAt"`
	Status   string             `json:"status"  bson:"status"`
}
