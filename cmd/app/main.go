package main

import (
	"log"
	"net/http"
)

func main() {
	// Minimal main for project scaffolding. Full server implemented in internal packages.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FamilyShare — initial scaffold"))
	})
	log.Println("Starting FamilyShare on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
