package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Парсим данные
func NextDate(now time.Time, date string, repeat string) (string, error) {
	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("incorrect date")
	}

	if repeat == "" {
		return "", errors.New("no repeat role")
	}

	parts := strings.Fields(repeat)

	// Обработка базовых правил
	switch parts[0] {
	case "d":
		if len(parts) < 2 {
			return "", errors.New("invalid rule format for 'd'")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("invalid number of days for 'd'")
		}

		// Проверяем, превышает ли новая дата текущую дату
		for {
			taskDate = taskDate.AddDate(0, 0, days)
			if taskDate.After(now) {
				return taskDate.Format("20060102"), nil
			}
		}

	case "y":
		for {
			taskDate = taskDate.AddDate(1, 0, 0)
			if taskDate.After(now) {
				return taskDate.Format("20060102"), nil
			}
		}

	default:
		return "", errors.New("incorrect format role")
	}
}

// NextDateHandler обрабатывает запрос для расчета следующей даты задачи
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из запроса
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Парсим параметр now
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "invalid 'now' parameter", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, nextDate)
}
