package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
	"todo_list/internal/model"
	"todo_list/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	taskService *service.TaskService
}

func NewHandler(taskService *service.TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

// CreateTask создает новую задачу.
// @Summary Создание новой задачи
// @Description Создает новую задачу
// @Accept json
// @Produce json
// @Param task body Task true "Задача"
// @Success 204 "No Content"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/todo-list/tasks [post]
func (s *Handler) CreateTask(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Printf("failed at ShouldBindJSON; error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = primitive.NewObjectID()
	task.Status = model.Active

	if len(task.Title) > 200 {
		log.Print("Title is too long")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long"})
		return
	}

	filter := bson.M{"title": task.Title, "activeAt": task.ActiveAt}
	err := s.taskService.DbManager.TaskCollection.FindOne(context.Background(), filter).Err()
	if err == nil {
		log.Print("Task with same title and activeAt already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "Task with same title and activeAt already exists"})
		return
	}

	_, err = s.taskService.DbManager.TaskCollection.InsertOne(context.Background(), task)
	if err != nil {
		log.Printf("failed at InsertOne; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateTask обновляет существующую задачу.
// @Summary Обновление задачи
// @Description Обновляет существующую задачу по идентификатору
// @Accept json
// @Produce json
// @Param id path string true "Идентификатор задачи"
// @Param task body Task true "Задача"
// @Success 204 "No Content"
// @Failure 400 "Bad Request"
// @Failure 404 "Not Found"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/todo-list/tasks/{id} [put]
// UpdateTask обрабатывает запрос на обновление существующей задачи.
func (s *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Printf("failed at ShouldBindJSON; error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(task.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long"})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Проверка существования задачи по ID
	filter := bson.M{"_id": objectId}
	err = s.taskService.DbManager.TaskCollection.FindOne(context.Background(), filter).Err()
	if err != nil {
		log.Printf("failed at FindOne; error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Обновление полей задачи
	update := bson.M{"$set": bson.M{
		"title":    task.Title,
		"activeAt": task.ActiveAt,
		// Другие поля, которые могут быть обновлены
	}}

	_, err = s.taskService.DbManager.TaskCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("failed at UpdateOne; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteTask удаляет задачу по идентификатору.
// @Summary Удаление задачи
// @Description Удаляет задачу по идентификатору
// @Produce json
// @Param id path string true "Идентификатор задачи"
// @Success 204 "No Content"
// @Failure 404 "Not Found"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/todo-list/tasks/{id} [delete]
// DeleteTask обрабатывает запрос на удаление задачи.
func (s *Handler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("failed at ObjectIDFromHex; error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	// Проверка существования задачи по ID
	filter := bson.M{"_id": objectId}
	err = s.taskService.DbManager.TaskCollection.FindOne(context.Background(), filter).Err()
	if err != nil {
		log.Printf("failed at FindOne; error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	_, err = s.taskService.DbManager.TaskCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("failed at DeleteOne; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkTaskDone помечает задачу как выполненную.
// @Summary Пометка задачи как выполненной
// @Description Помечает задачу как выполненную по идентификатору
// @Produce json
// @Param id path string true "Идентификатор задачи"
// @Success 204 "No Content"
// @Failure 404 "Not Found"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/todo-list/tasks/{id}/done [put]
// MarkTaskDone обрабатывает запрос на пометку задачи выполненой
func (s *Handler) MarkTaskDone(c *gin.Context) {
	id := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("failed at ObjectIDFromHex; error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Проверка существования задачи по ID
	filter := bson.M{"_id": objectId}
	err = s.taskService.DbManager.TaskCollection.FindOne(context.Background(), filter).Err()
	if err != nil {
		log.Printf("failed at FindOne; error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Обновление статуса на выполненный
	update := bson.M{"$set": bson.M{"status": model.Done}}

	_, err = s.taskService.DbManager.TaskCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("failed at UpdateOne; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTasksByStatus получает список задач по статусу.
// @Summary Список задач по статусу
// @Description Получает список задач по статусу (по умолчанию active)
// @Produce json
// @Param status query string false "Статус задачи (active или done)"
// @Success 200 {array} Task "Список задач"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/todo-list/tasks [get]
// GetTasksByStatus обрабатывает запрос на получение списка задач по статусу
func (s *Handler) GetTasksByStatus(c *gin.Context) {
	status := c.DefaultQuery("status", model.Active)

	filter := make([]bson.M, 0)

	if status == model.Active {
		filter = append(filter, bson.M{
			"activeAt": bson.M{"$lte": time.Now().Format("2006-01-02")},
		})
	}

	filter = append(filter, bson.M{
		"status": status,
	})

	cursor, err := s.taskService.DbManager.TaskCollection.Find(context.Background(), bson.M{"$and": filter})
	if err != nil {
		log.Printf("failed at Find; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var tasks []model.Task
	for cursor.Next(context.Background()) {
		var task model.Task
		cursor.Decode(&task)

		activeAt, err := time.Parse("2006-01-02", task.ActiveAt)
		if err != nil {
			log.Printf("failed at Parse; error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if activeAt.Weekday() == time.Sunday || activeAt.Weekday() == time.Saturday {
			task.Title = fmt.Sprintf("ВЫХОДНОЙ - %s", task.Title)
		}

		tasks = append(tasks, task)
	}

	// Сортировка задач по дате создания
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ActiveAt > tasks[j].ActiveAt
	})

	c.JSON(http.StatusOK, tasks)
}

func (s *Handler) GetTasks(c *gin.Context) {

	cursor, err := s.taskService.DbManager.TaskCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Printf("failed at Find; error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer cursor.Close(context.Background())

	var tasks []model.Task
	for cursor.Next(context.Background()) {
		var task model.Task
		cursor.Decode(&task)
		tasks = append(tasks, task)
	}

	// Сортировка задач по дате создания
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ActiveAt < tasks[j].ActiveAt
	})

	c.JSON(http.StatusOK, tasks)
}
