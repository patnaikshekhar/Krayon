package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func New(driver string, connectionString string) (*Database, error) {
	// if driver == "sqlite3" {
	// 	file, err := os.Create(connectionString) // Create SQLite file
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	file.Close()
	// }

	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}

	krayonDatabase := &Database{
		db: db,
	}

	createPluginTable := `
	CREATE TABLE IF NOT EXISTS plugins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT
	);`

	createPluginVersionsTable := `
	CREATE TABLE IF NOT EXISTS plugin_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		version TEXT NOT NULL,
		FOREIGN KEY(plugin_id) REFERENCES plugin(id)
	);`

	_, err = db.Exec(createPluginTable)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createPluginVersionsTable)
	if err != nil {
		return nil, err
	}

	return krayonDatabase, nil
}

func (krayonDatabase *Database) Close() error {
	return krayonDatabase.db.Close()
}
