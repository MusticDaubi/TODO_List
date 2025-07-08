package db

import "database/sql"

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = :id`
	_, err := DB.Exec(query, sql.Named("id", id))
	if err != nil {
		return err
	}
	return nil
}
