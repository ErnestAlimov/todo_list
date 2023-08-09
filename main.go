package main

import (
	"context"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	ID       string    `json:"id" bson:"_id"`
	Title    string    `json:"title"`
	ActiveAt time.Time `json:"activeAt"`
}

type TaskService struct {
	taskCollection *mongo.Collection
}

func NewTaskService(taskCollection *mongo.Collection) *TaskService {
	return &TaskService{
		taskCollection: taskCollection,
	}
}

// CreateTask обрабатывает запрос на создание новой задачи.
func (s *TaskService) CreateTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(task.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long"})
		return
	}

	// Проверка уникальности по полям title и activeAt
	existingTask := Task{}
	filter := bson.M{"title": task.Title, "activeAt": task.ActiveAt}
	err := s.taskCollection.FindOne(context.Background(), filter).Decode(&existingTask)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Task with same title and activeAt already exists"})
		return
	}

	_, err = s.taskCollection.InsertOne(context.Background(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateTask обрабатывает запрос на обновление существующей задачи.
func (s *TaskService) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(task.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long"})
		return
	}

	// Проверка существования задачи по ID
	filter := bson.M{"_id": id}
	existingTask := Task{}
	err := s.taskCollection.FindOne(context.Background(), filter).Decode(&existingTask)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Обновление полей задачи
	update := bson.M{"$set": bson.M{
		"title":    task.Title,
		"activeAt": task.ActiveAt,
		// Другие поля, которые могут быть обновлены
	}}

	_, err = s.taskCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteTask обрабатывает запрос на удаление задачи.
func (s *TaskService) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// Проверка существования задачи по ID
	filter := bson.M{"_id": id}
	existingTask := Task{}
	err := s.taskCollection.FindOne(context.Background(), filter).Decode(&existingTask)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	_, err = s.taskCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkTaskDone обрабатывает запрос на пометку задачи выполненой
func (s *TaskService) MarkTaskDone(c *gin.Context) {
	id := c.Param("id")

	// Проверка существования задачи по ID
	filter := bson.M{"_id": id}
	existingTask := Task{}
	err := s.taskCollection.FindOne(context.Background(), filter).Decode(&existingTask)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Обновление статуса на выполненный
	update := bson.M{"$set": bson.M{"status": true}}

	_, err = s.taskCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTasksByStatus обрабатывает запрос на получение списка задач по статусу
func (s *TaskService) GetTasksByStatus(c *gin.Context) {
	status := c.DefaultQuery("status", "active")
	filter := bson.M{}

	// Если статус active, фильтруем задачи по activeAt
	if status == "active" {
		today := time.Now()
		weekend := today.Weekday() == time.Saturday || today.Weekday() == time.Sunday
		if weekend {
			filter["title"] = bson.M{"$regex": ".*ВЫХОДНОЙ.*"}
		} else {
			filter["title"] = bson.M{"$regex": "^(?!.*ВЫХОДНОЙ.*)"}
		}
		filter["activeAt"] = bson.M{"$lte": today}
	}

	cursor, err := s.taskCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var tasks []Task
	for cursor.Next(context.Background()) {
		var task Task
		cursor.Decode(&task)
		tasks = append(tasks, task)
	}

	// Сортировка задач по дате создания
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ActiveAt.Before(tasks[j].ActiveAt)
	})

	c.JSON(http.StatusOK, tasks)
}

func main() {
	// Подключение к базе данных MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("todo_db")
	taskCollection := db.Collection("tasks")

	taskService := NewTaskService(taskCollection)
	// Создание роутера с использованием Gin
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		//Обработчики
		v1.POST("/todo-list/tasks", taskService.CreateTask)
		v1.PUT("/todo-list/tasks/:id", taskService.UpdateTask)
		v1.DELETE("/todo-list/tasks/:id", taskService.DeleteTask)
		v1.PUT("/todo-list/tasks/:id/done", taskService.MarkTaskDone)
		v1.GET("/todo-list/tasks", taskService.GetTasksByStatus)
	}
	//Запуск сервера
	router.Run(":8080")
}
