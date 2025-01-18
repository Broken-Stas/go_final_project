package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go_final_project/internal/next_date"
	"go_final_project/task"
)

func PostTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Формирование JSON-ответа
	var task task.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error": "title is not specified"}`, http.StatusBadRequest)
		return
	}

	// Проверка и обработка поля date

	if task.Date == "" || task.Date == "today" {
		task.Date = time.Now().Format(next_date.DateFormat)
	}
	parsedDate, err := time.Parse(next_date.DateFormat, task.Date)
	if err != nil {
		http.Error(w, `{"error": "incorrect date"}`, http.StatusBadRequest)
		return
	}

	if parsedDate.Before(time.Now()) {
		if task.Repeat != "" {
			nextDate, err := next_date.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		} else {
			// Если правила повторения нет - устанавливаем текущую дату
			task.Date = time.Now().Format(next_date.DateFormat)
		}
	}

	// Проверяем правило повторения
	if task.Repeat != "" {
		// Проверка на соответствие формату repeat, например: "d <целое число>", "w <целое число>", "m <целое число>"
		repeatPattern := `^([dwmy])\s?\d*$`
		matched, err := regexp.MatchString(repeatPattern, task.Repeat)
		if err != nil || !matched {
			http.Error(w, `{"error": "Incorrect repetition rule"}`, http.StatusBadRequest)
			return
		}
	}
	// Если дата задачи меньше текущей
	today := time.Now().Truncate(24 * time.Hour)
	if parsedDate.Before(today) {
		if task.Repeat == "" {
			// Если правило повторения не указано, устанавливаем текущую дату.
			parsedDate = today
		} else {
			// Если нет, вычисляем следующую допустимую дату
			nextDate, err := next_date.NextDate(today, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error": "Error calculating the next date"}`, http.StatusBadRequest)
				return
			}
			parsedNextDate, err := time.Parse(next_date.DateFormat, nextDate)
			if err != nil {
				http.Error(w, `{"error": "Date conversion error"}`, http.StatusBadRequest)
				return
			}
			parsedDate = parsedNextDate
		}
	}

	// Преобразуем дату обратно в строку
	task.Date = parsedDate.Format(next_date.DateFormat)

	// Проверяем правило повторения на корректность
	validRepeat := false
	validFormats := []string{"d", "w", "m", "y"}

	for _, validFormat := range validFormats {
		if strings.HasPrefix(task.Repeat, validFormat) {
			validRepeat = true
			break
		}
	}

	if !validRepeat && task.Repeat != "" {
		http.Error(w, `{"error": "Incorrect repetition rule"}`, http.StatusBadRequest)
		return
	}

	// Добавление задачи в БД
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Error when adding a task to the database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Получаем id задачи
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "task ID error: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Возвращаем id добавленной задачи
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}
