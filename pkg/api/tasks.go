package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"final_project/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}
type ErrorResp struct {
	Error string `json:"error"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	flag := false
	check := strings.TrimSpace(search)

	if len(search) == 10 && check[2] == '.' && check[5] == '.' {
		flag = true
	}
	if search != "" {
		if flag {
			t, err := time.Parse("02.01.2006", search)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResp{Error: err.Error()})
				return
			}
			search = t.Format(Format)
		}
	}

	tasks, err := db.Tasks(30, flag, search)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{Error: err.Error()})
		return
	}
	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TasksResp{Tasks: tasks})
}

func getTask(id string) (*db.Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id`
	var t db.Task
	err := db.DB.QueryRow(query, sql.Named("id", id)).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found (id: %d)", id)
		}
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &t, nil
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{"Failed to read request body"})
		return
	}
	defer r.Body.Close()

	var t db.Task
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{"Invalid JSON format"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if t.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{"Task ID is required"})
		return
	}

	_, err = getTask(t.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResp{err.Error()})
		return
	}

	if t.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{"Title is required"})
		return
	}

	if err = CheckDate(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{err.Error()})
		return
	}

	query := `UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id`
	res, err := db.DB.Exec(query, sql.Named("date", t.Date), sql.Named("title", t.Title), sql.Named("comment", t.Comment), sql.Named("repeat", t.Repeat), sql.Named("id", t.ID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResp{err.Error()})
		return
	}

	checkRows, err := res.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResp{err.Error()})
		return
	}

	if checkRows == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResp{"task not found"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{}{})
	return
}
