package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

func writeContent(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), os.ModePerm)
}

func handle(w http.ResponseWriter, r *http.Request) {
	err := writeContent(path.Join(".", r.URL.Path), r.FormValue("content"))
	if err == nil {
		w.Header().Set("Location", path.Join(r.URL.Path, "success"))
	} else {
		log.Println(err)
		w.Header().Set("Location", path.Join(r.URL.Path, "failure"))
	}

	w.WriteHeader(http.StatusSeeOther)
}

func main() {
	log.Fatal(http.ListenAndServe(":9092", http.HandlerFunc(handle)))
}
