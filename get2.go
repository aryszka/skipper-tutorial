package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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

	// apply the new uppercase feature only when old version is not forced
	if c, err := r.Cookie("version"); err != nil || c.Value != "old" {
		content = strings.ToUpper(content)
	}

	page.Execute(w, content)
}

func main() {
	log.Fatal(http.ListenAndServe(":9093", http.HandlerFunc(handle)))
}
