package handler

import (
	"database/sql"
	"fmt"
	"net/http"
)

func DeleteTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем ID задачи из параметров запроса.
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		// Если ID не найден, возвращаем ошибку.
		http.Error(w, `{"error":"missing task ID"}`, http.StatusBadRequest)
		return
	}

	// Удаляем задачу из БД.
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
	if err != nil {
		// Если возникла ошибка при выполнении запроса, возвращаем ошибку сервера.
		http.Error(w, `{"error": "Error deleting task"}`, http.StatusInternalServerError)
		return
	}

	// Проверяем, была ли удалена задача.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Если не удалось получить количество затронутых строк, возвращаем ошибку сервера.
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		// Если ни одной строки не было удалено, возвращаем ошибку "task not found".
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
