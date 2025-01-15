package main

import (
	"go_final_project/db"
	"go_final_project/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	webDir := "./web"

	// Получение порта из переменной окружения или значения по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Инициализация БД
	db, err := db.InitDB("scheduler.db")
	if err != nil {
		log.Fatalf("Ошибка инициализации DB: %v", err)
	}

	// Настройка маршрутов
	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)
	http.HandleFunc("/api/nextdate", handler.NextDateHandler)
	http.HandleFunc("/api/task", handler.TaskHandler(db))
	http.HandleFunc("/api/tasks", handler.GetTasksHandler(db))
	http.HandleFunc("/api/task/done", handler.DoneTaskHandler(db))

	// Запуск сервера
	log.Printf("Сервер запущен на порту %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %s", err)
	}
}
