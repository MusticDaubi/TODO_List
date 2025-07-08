package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"final_project/pkg/db"
)

type TaskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	var response TaskResponse
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Error = "Invalid JSON: " + err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	if task.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		response.Error = "Title is required"
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := CheckDate(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Error = "Database error: " + err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	response.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func CheckDate(task *db.Task) error {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	if task.Date == "" {
		task.Date = now.Format(Format)
	}

	t, err := time.Parse(Format, task.Date)
	if err != nil {
		return errors.New("invalid date format")
	}
	t = t.UTC().Truncate(24 * time.Hour)

	if t.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(Format)
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
		}
	}
	if task.Repeat != "" {
		_, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	return nil
}
