package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"io"
)

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := os.Create(path.Join("data", r.URL.Path))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.Copy(f, r.Body)
}

func main() {
	log.Fatal(http.ListenAndServe(":9092", http.HandlerFunc(handle)))
}
