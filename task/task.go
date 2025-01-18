package task

import (
	"database/sql"
	"fmt"
)

// Task представляет задачу в базе данных.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

const limit = 50

// GetTasks возвращает все задачи из БД
func GetTasks(db *sql.DB) ([]Task, error) {
	query := fmt.Sprintf("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT %d", limit)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("string scanning error: %v", err)
		}
		tasks = append(tasks, task)
	}
	// Добавляем проверку на ошибку после цикла
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after for loop: %v", err)
	}

	return tasks, nil
}

// GetTasksBySearch ищет задачи по заголовку или комментарию.
func GetTasksBySearch(db *sql.DB, search string) ([]Task, error) {
	query := fmt.Sprintf("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT %d", limit)
	rows, err := db.Query(query, "%"+search+"%", "%"+search+"%")
	if err != nil {
		return nil, fmt.Errorf("request execution error: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("string scanning error: %v", err)
		}
		tasks = append(tasks, task)
	}
	// Добавляем проверку на ошибку после цикла
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after for loop: %v", err)
	}

	return tasks, nil
}

// GetTasksByDate возвращает задачи на указанную дату.
func GetTasksByDate(db *sql.DB, date string) ([]Task, error) {
	query := fmt.Sprintf("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT %d", limit)
	rows, err := db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("string scanning error: %v", err)
		}
		tasks = append(tasks, task)
	}
	// Добавляем проверку на ошибку после цикла
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after for loop: %v", err)
	}

	return tasks, nil
}
