package db

import "log"

type Plugin struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PluginVersion struct {
	ID       int    `json:"id"`
	PluginID int    `json:"plugin_id"`
	Version  string `json:"version"`
}

func (kdb *Database) InsertPluginWithVersion(name, description, version string) (*Plugin, error) {

	tx, err := kdb.db.Begin()
	if err != nil {
		return nil, err
	}

	// Query if a plugin with the same name already exists
	var pluginID int64
	err = tx.QueryRow("SELECT id FROM plugins WHERE name = ?", name).Scan(&pluginID)
	if err != nil {
		result, err := tx.Exec("INSERT INTO plugins (name, description) VALUES (?, ?)", name, description)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		pluginID, _ = result.LastInsertId()
	}

	// Query if a plugin version already exists
	var pluginVersionID int
	err = tx.QueryRow("SELECT id FROM plugin_versions WHERE plugin_id = ? AND version = ?", pluginID, version).Scan(&pluginVersionID)
	if err != nil {
		_, err = tx.Exec("INSERT INTO plugin_versions (plugin_id, version) VALUES (?, ?)", pluginID, version)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit the transaction if plugin insertion is successful
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &Plugin{ID: int(pluginID), Name: name, Description: description}, nil
}

func (kdb *Database) GetAllPlugins() ([]*Plugin, error) {
	rows, err := kdb.db.Query("SELECT id, name, description FROM plugins")
	if err != nil {
		return nil, err
	}

	var pg []*Plugin

	for rows.Next() {
		log.Printf("Plugins loop")
		p := &Plugin{}
		err = rows.Scan(&p.ID, &p.Name, &p.Description)
		if err != nil {
			return nil, err
		}

		pg = append(pg, p)
	}
	return pg, nil
}

func (kdb *Database) GetPluginVersions(id string) ([]*PluginVersion, error) {
	rows, err := kdb.db.Query("SELECT id, plugin_id, version FROM plugin_versions WHERE plugin_id=?", id)
	if err != nil {
		return nil, err
	}

	var pv []*PluginVersion

	for rows.Next() {
		p := &PluginVersion{}
		err = rows.Scan(&p.ID, &p.PluginID, &p.Version)
		if err != nil {
			return nil, err
		}

		pv = append(pv, p)
	}
	return pv, nil
}
