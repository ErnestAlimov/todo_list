# todo_list
Это микросервис для управления списком задач (Todo List) с использованием фреймворка Gin и базы данных MongoDB. Микросервис предоставляет функциональность создания, обновления, удаления, пометки как выполненной и просмотра задач.

Для запуска в приложении запустите docker-compose build и docker-compose up
API Endpoints

Создание задачи
POST /api/v1/todo-list/tasks
Создает новую задачу.

Обновление задачи
PUT /api/v1/todo-list/tasks/{id}
Обновляет существующую задачу по идентификатору.

Удаление задачи
DELETE /api/v1/todo-list/tasks/{id}
Удаляет задачу по идентификатору.

Пометка задачи выполненной
PUT /api/v1/todo-list/tasks/{id}/done
Помечает задачу как выполненную.

Список задач по статусу
GET /api/v1/todo-list/tasks?status=active
Получает список задач по статусу (по умолчанию active).

