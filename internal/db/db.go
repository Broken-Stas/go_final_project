package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Инициализация БД
func InitDB(dbFile string) (*sql.DB, error) {
	// Проверка наличии БД
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	// Подключение к БД, если нет - создаем
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if install {
		log.Println("Creating a database.")
		createTableSQL := `
            CREATE TABLE scheduler (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                date CHAR(8) NOT NULL DEFAULT "",
                title VARCHAR(128) NOT NULL DEFAULT
                comment NOT NULL DEFAULT "",
                repeat VARCHAR(128) NOT NULL
            );
            CREATE INDEX idx_date ON scheduler (date);
        `
		if _, err := db.Exec(createTableSQL); err != nil {
			return nil, err
		}
		log.Println("The scheduler table has been created.")
	}
	return db, nil
}
