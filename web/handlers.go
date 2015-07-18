package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

var (
	cachedTemplates = make(map[string]*template.Template)
)

func New(templateDir string) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		t := getTemplate("list_jobs", templateDir)
		must(t.Execute(w, nil))
	}).Methods("GET")

	router.HandleFunc("/jobs/new", func(w http.ResponseWriter, r *http.Request) {
		t := getTemplate("create_job", templateDir)
		must(t.Execute(w, nil))
	}).Methods("GET")

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/jobs/42/output", 302)
	}).Methods("POST")

	router.HandleFunc("/jobs/{id}/output", func(w http.ResponseWriter, r *http.Request) {
		t := getTemplate("job_output", templateDir)
		must(t.Execute(w, nil))
	}).Methods("GET")

	return router
}

func getTemplate(name, templateDir string) *template.Template {
	templatePath := filepath.Join(templateDir, fmt.Sprintf("%s.html", name))
	if _, ok := cachedTemplates[templatePath]; !ok {
		cachedTemplates[templatePath] = template.Must(template.ParseFiles(templatePath))
	}
	return cachedTemplates[templatePath]
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
