package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	var id int64
	if DB == nil {
		return 0, errors.New("db not initialized")
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := DB.Exec(query, sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, errors.New("add task failed: " + err.Error())
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, errors.New("get ID failed: " + err.Error())
	}
	return id, err
}

func Tasks(limit int, flag bool, search ...string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	searchVal := ""
	if len(search) > 0 {
		searchVal = search[0]
	}

	if searchVal != "" {
		switch flag {
		case true:
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = :date LIMIT ?`
			rows, err = DB.Query(query, sql.Named("date", searchVal), limit)
		case false:
			searchVal = "%" + searchVal + "%"
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit`
			rows, err = DB.Query(query, sql.Named("search", searchVal), sql.Named("limit", limit))
		}
	} else {
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}

	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}
