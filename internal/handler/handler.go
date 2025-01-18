package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go_final_project/task"
)

// GetTasksHandler обрабатывает запросы на получение ближайших задач.
func GetTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем параметр search из строки запроса
		search := r.URL.Query().Get("search")

		// Если параметр search не пустой, делаем поиск
		var tasks []task.Task
		var err error
		if search != "" {
			// Проверяем, соответствует ли search формату даты
			if _, parseErr := time.Parse("02.01.2006", search); parseErr == nil {
				search = strings.Replace(search, ".", "", -1) // Преобразуем формат
				tasks, err = task.GetTasksByDate(db, search)
			} else {
				tasks, err = task.GetTasksBySearch(db, search)
			}

		} else {
			// Если search не указан, получаем все задачи
			tasks, err = task.GetTasks(db)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Error getting tasks from the database: %v"}`, err), http.StatusInternalServerError)
			return
		}

		// Формируем JSON-ответ
		response := struct {
			Tasks []task.Task `json:"tasks"`
		}{
			Tasks: tasks,
		}

		// Если задач нет, возвращаем пустой список
		if len(tasks) == 0 {
			response.Tasks = []task.Task{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
