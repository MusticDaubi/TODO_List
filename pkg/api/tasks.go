package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"final_project/pkg/db"
)

const limit = 30

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
				sendError(w, http.StatusBadRequest, err.Error())
				return
			}
			search = t.Format(Format)
		}
	}

	tasks, err := db.Tasks(limit, flag, search)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	sendResponse(w, http.StatusOK, TasksResp{Tasks: tasks})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var t db.Task
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if t.ID == "" {
		sendError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	_, err = db.GetTask(t.ID)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	if t.Title == "" {
		sendError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if err = CheckDate(&t); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := db.Update(t)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	checkRows, err := res.RowsAffected()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if checkRows == 0 {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}

	sendResponse(w, http.StatusOK, struct{}{})
	return
}
