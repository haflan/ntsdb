package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "embed"
	"github.com/haflan/ntsdb"
)

//go:embed index.html
var index []byte

type notesrv struct {
	db ntsdb.NTSDB
}

func NewNoteServer(db ntsdb.NTSDB) http.Handler {
	s := &notesrv{db}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write(index) })
	mux.HandleFunc("/notes", s.handleNotes)
	mux.HandleFunc("/groups", s.handleGroups)
	return mux
}

func (n notesrv) handleNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := n.db.GetNotes()
	if err != nil {
		log.Printf("failed to get notes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		log.Printf("failed to write notes: %v", err)
	}
}

func (n notesrv) handleGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := n.db.GetNoteGroups()
	if err != nil {
		log.Printf("failed to get note groups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		log.Printf("failed to write note groups: %v", err)
	}
}
