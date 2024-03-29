package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
)

type config struct {
	authKey string
	master  bool
}

type uploadResponse struct {
	Status string `json:"status"`
	File   string `json:"file"`
}

func StartAsMaster(network string, apiKey string, store Store) {
	m := martini.Classic()
	config := &config{
		master:  true,
		authKey: apiKey,
	}

	m.Map(log.New(ioutil.Discard, "", 0))
	m.Map(store)
	m.Map(config)

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/upload", uploadRaw)
	m.Get("/:id", downloadRaw)

	logger.Info("Starting as master node at %s", network)
	logger.Fatal(http.ListenAndServe(network, m))
}

func downloadRaw(w http.ResponseWriter, r *http.Request, params martini.Params, store Store) {
	if filePath, error := store.GetFilePath(params["id"]); error == nil {
		http.ServeFile(w, r, filePath)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func uploadRaw(w http.ResponseWriter, r *http.Request, store Store, config *config) {
	r.ParseMultipartForm(64 << 20)

	apikey := r.FormValue("apiKey")
	if apikey != config.authKey {
		http.Error(w, "Api key is required", http.StatusUnauthorized)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id, err := store.StoreFile(file)
	if err != nil {
		logger.Error("Can't store file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	json, err := json.Marshal(&uploadResponse{Status: "ok", File: id})
	if err != nil {
		logger.Error("Can't store file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
