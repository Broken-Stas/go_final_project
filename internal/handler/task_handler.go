package handler

import (
	"database/sql"
	"net/http"
)

func TaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetTask(db, w, r)
			return
		}
		if r.Method == http.MethodPost {
			PostTask(db, w, r)
			return
		}
		if r.Method == http.MethodPut {
			PutTask(db, w, r)
			return
		}
		if r.Method == http.MethodDelete {
			DeleteTask(db, w, r)
			return
		}
		http.Error(w, `{"error": "method is not supported"}`, http.StatusMethodNotAllowed)
	}
}
