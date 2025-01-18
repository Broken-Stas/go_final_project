package main

import (
	"log"
	"net/http"
	"os"

	"go_final_project/internal/db"
	"go_final_project/internal/handler"
	"go_final_project/internal/next_date"
)

func main() {
	webDir := "./web"

	// Получение порта из переменной окружения или значения по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Инициализация БД
	db, err := db.InitDB("./scheduler.db")
	if err != nil {
		log.Fatalf("DB initialization error: %v", err)
	}

	// Настройка маршрутов
	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)
	http.HandleFunc("/api/nextdate", next_date.NextDateHandler)
	http.HandleFunc("/api/task", handler.TaskHandler(db))
	http.HandleFunc("/api/tasks", handler.GetTasksHandler(db))
	http.HandleFunc("/api/task/done", handler.DoneTaskHandler(db))

	// Запуск сервера
	log.Printf("The server is running on the port: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server startup error: %s", err)
	}
}
