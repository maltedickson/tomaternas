package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New() (*DB, error) {
	db, err := sql.Open("sqlite", "./data/recipes.db")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// _, err = db.Exec(schema)
	// if err != nil {
	// 	return nil, err
	// }

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
