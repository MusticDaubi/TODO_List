package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(128),
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

func Init(dbFile string) error {
	var install bool
	_, err := os.Stat(dbFile)
	if err != nil {
		install = true
	}
	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("error opening database file: %v", err)
	}

	if install {
		if _, err := DB.Exec(schema); err != nil {
			return fmt.Errorf("error installing database schema: %v", err)
		}
	}
	return nil
}
