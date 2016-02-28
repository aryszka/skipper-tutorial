package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

var page *template.Template

func init() {
	t, err := template.ParseFiles("template.html")
	if err != nil {
		panic(err)
	}

	page = t
}

func readContent(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	s := string(b)
	if s == "" {
		s = "[no content]"
	}

	return s, nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	content, err := readContent(path.Join(".", r.URL.Path))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	page.Execute(w, content)
}

func main() {
	log.Fatal(http.ListenAndServe(":9091", http.HandlerFunc(handle)))
}
