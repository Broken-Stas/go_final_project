package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go_final_project/internal/next_date"
	"go_final_project/task"
)

func PutTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var task task.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, `{"error": "Error deserializing JSON"}`, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error": "task title is not specified"}`, http.StatusBadRequest)
		return
	}

	if task.Comment == "" {
		http.Error(w, `{"error": "task comment is not specified"}`, http.StatusBadRequest)
		return
	}
	if task.Repeat == "" {
		http.Error(w, `{"error": "Task repetition is not specified"}`, http.StatusBadRequest)
		return
	}

	if task.Date == "" || task.Date == "today" {
		task.Date = time.Now().Format(next_date.DateFormat)
	}

	// Проверяем, существует ли задача с данным id
	query := "SELECT id FROM scheduler WHERE id = ?"
	var existingID int
	err = db.QueryRow(query, task.ID).Scan(&existingID)
	if err != nil || existingID == 0 {
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}

	// Обновляем задачу в БД
	updateQuery := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "task update error: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
