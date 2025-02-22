package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/andrewgonzales/go-eggwalker/fuguemax"
)

var (
	leftDoc  = fuguemax.NewDoc("left")
	rightDoc = fuguemax.NewDoc("right")
	mu       sync.Mutex
)

func StartServer() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/send-right", sendRight)
	http.HandleFunc("/send-left", sendLeft)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/reset", reset)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Listening on port 80...")
	http.ListenAndServe(":80", nil)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		LeftText  string
		RightText string
	}{
		LeftText:  leftDoc.StringContent(),
		RightText: rightDoc.StringContent(),
	}
	tmpl.Execute(w, data)
}

func sendRight(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		Text     string `json:"text"`
		Position uint64 `json:"position"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	leftDoc.LocalInsertText(req.Text, req.Position)
	leftDoc.MergeInto(&rightDoc)
	rightDoc.MergeInto(&leftDoc)

	resp := map[string]string{
		"left":  leftDoc.StringContent(),
		"right": rightDoc.StringContent(),
	}
	json.NewEncoder(w).Encode(resp)
}

func sendLeft(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		Text     string `json:"text"`
		Position uint64 `json:"position"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	rightDoc.LocalInsertText(req.Text, req.Position)
	rightDoc.MergeInto(&leftDoc)
	leftDoc.MergeInto(&rightDoc)

	resp := map[string]string{
		"left":  leftDoc.StringContent(),
		"right": rightDoc.StringContent(),
	}
	json.NewEncoder(w).Encode(resp)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		Agent        string `json:"agent"`
		Position     uint64 `json:"position"`
		NumDeletions int    `json:"numDeletions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Perform the deletion on both documents
	if req.Agent == "left" {
		leftDoc.LocalDelete(req.Position, req.NumDeletions)
		leftDoc.MergeInto(&rightDoc)
		rightDoc.MergeInto(&leftDoc)
	} else {
		rightDoc.LocalDelete(req.Position, req.NumDeletions)
		rightDoc.MergeInto(&leftDoc)
		leftDoc.MergeInto(&rightDoc)
	}

	resp := map[string]string{
		"left":  leftDoc.StringContent(),
		"right": rightDoc.StringContent(),
	}
	json.NewEncoder(w).Encode(resp)
}

func reset(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	mu.Lock()
	defer mu.Unlock()

	leftDoc = fuguemax.NewDoc("agent1")
	rightDoc = fuguemax.NewDoc("agent2")

	resp := map[string]string{
		"left":  leftDoc.StringContent(),
		"right": rightDoc.StringContent(),
	}
	json.NewEncoder(w).Encode(resp)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	(*w).Header().Set("Access-Control-Allow-Headers", "content-type, Authorization")
}
