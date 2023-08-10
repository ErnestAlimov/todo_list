package service

import (
	dbmanager "todo_list/internal/db-manager"
)

type TaskService struct {
	DbManager *dbmanager.DbManager
}

func NewTaskService(dbManager *dbmanager.DbManager) *TaskService {
	return &TaskService{
		DbManager: dbManager,
	}
}
