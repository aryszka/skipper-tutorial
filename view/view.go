package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "strings"
    "path"
)

var page *template.Template

func init() {
	t, err := template.ParseFiles("view/view.html")
	if err != nil {
		panic(err)
	}

	page = t
}

func loadContent(p string) ([]byte, error) {
    return ioutil.ReadFile(path.Join("data", p))
}

func wantsText(r *http.Request) bool {
    return strings.HasPrefix(r.Header.Get("Accept"), "text/plain")
}

func notFound(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func handle(w http.ResponseWriter, r *http.Request) {
	content, err := loadContent(r.URL.Path)
    if os.IsNotExist(err) {
        notFound(w, r)
        return
    }

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

    if wantsText(r) {
        w.Write(content)
        return
    }

	page.Execute(w, map[string]string{
		"Title": r.URL.Path[1:],
		"Content": string(content)})
}

func main() {
	log.Fatal(http.ListenAndServe(":9091", http.HandlerFunc(handle)))
}
