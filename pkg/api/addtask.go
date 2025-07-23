package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
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

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if task.Title == "" {
		sendError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if err := CheckDate(&task); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	response.ID = id
	sendResponse(w, http.StatusCreated, response)

}

func sendError(w http.ResponseWriter, status int, message string) {
	response := TaskResponse{Error: message}
	sendResponse(w, status, response)
}

func sendResponse(w http.ResponseWriter, status int, data interface{}) {
	var buf bytes.Buffer
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func CheckDate(task *db.Task) error {
	now := time.Now().Local()

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
