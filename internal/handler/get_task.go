package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"go_final_project/task"
)

func GetTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	idString := r.URL.Query().Get("id")
	if idString == "" {
		http.Error(w, `{"error": " id query parameter is required"}`, http.StatusMethodNotAllowed)
		return

	}
	var task task.Task
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"

	id, _ := strconv.Atoi(idString)
	db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if task.ID == "" {
		http.Error(w, `{"error": "task not found"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}
