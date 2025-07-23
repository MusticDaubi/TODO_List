package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"final_project/pkg/db"

	"github.com/golang-jwt/jwt/v5"
)

var (
	Format       = "20060102"
	todoPassword string
)

func Init(password string) {
	todoPassword = password

	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(taskDoneHandler))
	http.HandleFunc("/api/signin", authHandler)
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	if now == "" {
		now = time.Now().Format(Format)
	}
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	timeNow, err := time.Parse(Format, now)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid 'now' date format")
		return
	}
	result, err := NextDate(timeNow, date, repeat)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	type NextDateResponse struct {
		Date string `json:"date"`
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(result))
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getSingleTaskHandler(w, r)
	case http.MethodPut:
		updateTask(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	id := r.URL.Query().Get("id")
	t, err := db.GetTask(id)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if t.Repeat == "" {
		if err = db.DeleteTask(id); err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sendResponse(w, http.StatusOK, struct{}{})
		return
	}
	now := time.Now()
	nextDate, err := NextDate(now, t.Date, t.Repeat)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = db.UpdateTask(nextDate, t.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, struct{}{})
	return
}

func getSingleTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			sendError(w, http.StatusNotFound, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendResponse(w, http.StatusOK, task)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	_, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			sendError(w, http.StatusNotFound, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, struct{}{})
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	type RequestData struct {
		Password string `json:"password"`
	}

	var reqData RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if reqData.Password == "" {
		sendError(w, http.StatusBadRequest, "password is required")
		return
	}

	if todoPassword != reqData.Password {
		sendError(w, http.StatusBadRequest, "incorrect password")
		return
	}

	pass := []byte(todoPassword)
	hash := sha256.Sum256(pass)
	hashedPass := hex.EncodeToString(hash[:])

	claims := jwt.MapClaims{
		"password_hash": hashedPass,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString(pass)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"token": signedToken})
	return
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			sendError(w, http.StatusUnauthorized, "missing token")
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(todoPassword), nil
		})

		if err != nil {
			sendError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
			return
		}

		if !token.Valid {
			sendError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			sendError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		tokenHash, ok := claims["password_hash"].(string)
		if !ok {
			sendError(w, http.StatusUnauthorized, "missing password hash")
			return
		}

		currentHashBytes := sha256.Sum256([]byte(todoPassword))
		currentHash := hex.EncodeToString(currentHashBytes[:])
		if currentHash != tokenHash {
			sendError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next(w, r)
	}
}
