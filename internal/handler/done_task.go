package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"go_final_project/internal/next_date"
)

func DoneTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method is not supported"}`, http.StatusMethodNotAllowed)
			return
		}

		// Получаем ID задачи из параметров запроса.
		taskID := r.URL.Query().Get("id")
		if taskID == "" {
			http.Error(w, `{"error":"missing task ID"}`, http.StatusBadRequest)
			return
		}

		// Получаем задачу из базы данных.
		var task struct {
			ID     int
			Date   string
			Repeat string
		}
		query := "SELECT id, date, repeat FROM scheduler WHERE id = ?"
		err := db.QueryRow(query, taskID).Scan(&task.ID, &task.Date, &task.Repeat)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
			} else {
				http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			}
			return
		}

		// Обработка задачи в зависимости от наличия правила повторения.
		if task.Repeat == "" {
			// Если задача одноразовая, удаляем её.
			_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
			if err != nil {
				http.Error(w, `{"error":"failed to delete task"}`, http.StatusInternalServerError)
				return
			}
		} else {
			// Если задача повторяющаяся, рассчитываем следующую дату.
			if task.Repeat != "" {
				nextDate, err := next_date.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusBadRequest)
					return
				}
				task.Date = nextDate
			}

			// Обновляем дату задачи в БД.
			_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", task.Date, task.ID)
			if err != nil {
				http.Error(w, `{"error":"failed to update task"}`, http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}
