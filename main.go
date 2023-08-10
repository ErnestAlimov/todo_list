package main

import (
	dbmanager "todo_list/internal/db-manager"
	"todo_list/internal/handlers"
	"todo_list/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Подключение к базе данных MongoDB

	dbManager := dbmanager.NewDbManager("localhost", "27017", "test", "test")
	taskService := service.NewTaskService(dbManager)

	handlers := handlers.NewHandler(taskService)

	// Создание роутера с использованием Gin
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		//Обработчики
		v1.POST("/todo-list/tasks", handlers.CreateTask)
		v1.PUT("/todo-list/tasks/:id", handlers.UpdateTask)
		v1.DELETE("/todo-list/tasks/:id", handlers.DeleteTask)
		v1.PUT("/todo-list/tasks/:id/done", handlers.MarkTaskDone)
		v1.GET("/todo-list/tasks", handlers.GetTasksByStatus)
	}

	//Запуск сервера
	router.Run(":8080")
}
