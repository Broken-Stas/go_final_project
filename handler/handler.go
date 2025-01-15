package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/task"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func TaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getTask(db, w, r)
			return
		}
		if r.Method == http.MethodPost {
			postTask(db, w, r)
			return
		}
		if r.Method == http.MethodPut {
			putTask(db, w, r)
			return
		}
		if r.Method == http.MethodDelete {
			deleteTask(db, w, r)
			return
		}
		http.Error(w, `{"error": "method is not supported"}`, http.StatusMethodNotAllowed)
	}
}

func getTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

func postTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
		task.Date = time.Now().Format("20060102")
	}
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error": "incorrect date"}`, http.StatusBadRequest)
		return
	}

	if parsedDate.Before(time.Now()) {
		if task.Repeat != "" {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		} else {
			// Если правила повторения нет - устанавливаем текущую дату
			task.Date = time.Now().Format("20060102")
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
			nextDate, err := NextDate(today, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error": "Error calculating the next date"}`, http.StatusBadRequest)
				return
			}
			parsedNextDate, err := time.Parse("20060102", nextDate)
			if err != nil {
				http.Error(w, `{"error": "Date conversion error"}`, http.StatusBadRequest)
				return
			}
			parsedDate = parsedNextDate
		}
	}

	// Преобразуем дату обратно в строку
	task.Date = parsedDate.Format("20060102")

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
func putTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
		task.Date = time.Now().Format("20060102")
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
				nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
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

func deleteTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
