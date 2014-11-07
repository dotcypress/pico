package main

import (
	"encoding/json"
	"net/http"
)

type uploadResponse struct {
	Status string `json:"status"`
	File   string `json:"file"`
}

type handler func(w http.ResponseWriter, r *http.Request)

var authKey string

func Serve(network string, apiKey string, isMaster bool) {
	authKey = apiKey
	http.Handle("/", http.FileServer(http.Dir(store.GetPath())))
	if isMaster {
		log.Info("Starting as master node at: %s with store: %s", network, store)
		http.HandleFunc("/upload", auth(postOnly(upload)))
	} else {
		log.Info("Starting as slave node at: %s with store: %s", network, store)
	}
	log.Fatal(http.ListenAndServe(network, nil))
}

func auth(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header["Authorization"]
		if len(apiKey) == 0 || authKey != apiKey[0] {
			log.Error("Authorization failed", r)
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

func postOnly(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			h(w, r)
			return
		}
		log.Error(r.Method+" not allowed", r)
		http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	id, err := store.StoreFile(file)
	if err != nil {
		log.Error("Can't store file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	json, err := json.Marshal(&uploadResponse{Status: "ok", File: id})
	if err != nil {
		log.Error("Can't store file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
