package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	database "krayon/internal/db"
	"krayon/internal/storage"

	"github.com/urfave/cli/v2"
)

func PluginsServe(ctx *cli.Context) error {
	port := ctx.Int("port")

	db, err := database.New(ctx.String("driver"), ctx.String("connection-string"))
	if err != nil {
		return err
	}
	defer db.Close()

	storage, err := storage.NewStorage(ctx.String("storage-type"), ctx.String("bucket"))
	if err != nil {
		return err
	}

	http.HandleFunc("/api/v1/plugins/{id}/versions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getPluginVersions(db)(w, r)
		} else if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/api/v1/plugins/{pluginName}/versions/{pluginVersion}/platforms/{platformName}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			downloadPlugin(storage)(w, r)
		} else if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/api/v1/plugins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getAllPlugins(db)(w, r)
		} else if r.Method == http.MethodPost {
			insertPluginAndVersionHandler(db, storage)(w, r)
		}
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return err
	}

	return nil
}

type PluginInsertRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type PluginInsertResponse struct {
	ID int `json:"id"`
}

func downloadPlugin(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("pluginName")
		version := r.PathValue("pluginVersion")
		platformName := r.PathValue("platformName")
		binary, err := storage.Download(r.Context(), name, version, platformName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer binary.Close()

		_, err = io.Copy(w, binary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getPluginVersions(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		plugins, err := db.GetPluginVersions(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(plugins)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getAllPlugins(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugins, err := db.GetAllPlugins()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(plugins)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func insertPluginAndVersionHandler(db *database.Database, storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p PluginInsertRequest

		// Parse multipart form data
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get plugin details
		p.Name = r.FormValue("name")
		p.Description = r.FormValue("description")
		p.Version = r.FormValue("version")

		insertedPlugin, err := db.InsertPluginWithVersion(p.Name, p.Description, p.Version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = uploadFiles(storage, r, p.Name, p.Version, "mac", "windows", "linux")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(insertedPlugin)
	}
}

func uploadFiles(storage storage.Storage, r *http.Request, name, version string, arcs ...string) error {
	for _, arc := range arcs {
		file, _, err := r.FormFile("binary_" + arc)
		if err != nil {
			return err
		}
		defer file.Close()
		err = storage.Store(context.Background(), name, version, arc, file)
		if err != nil {
			return err
		}
	}

	return nil
}
