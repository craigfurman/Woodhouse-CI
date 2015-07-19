package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router

	templates map[string]*template.Template
}

func New(templateDir string) *Handler {
	templates := parseTemplates(templateDir)
	router := mux.NewRouter()

	handler := &Handler{
		Router:    router,
		templates: templates,
	}

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("list_jobs", w)
	}).Methods("GET")

	router.HandleFunc("/jobs/new", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("create_job", w)
	}).Methods("GET")

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/jobs/42/output", 302)
	}).Methods("POST")

	router.HandleFunc("/jobs/{id}/output", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("job_output", w)
	}).Methods("GET")

	return handler
}

func (h Handler) renderTemplate(name string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	must(h.templates[name].Execute(w, nil))
}

func parseTemplates(templateDir string) map[string]*template.Template {
	templates := make(map[string]*template.Template)

	layout := filepath.Join(templateDir, "layouts", "layout.html")
	views, err := filepath.Glob(fmt.Sprintf("%s/views/*.html", templateDir))
	must(err)

	for _, view := range views {
		viewName := strings.Split(filepath.Base(view), ".")[0]
		templates[viewName] = template.Must(template.ParseFiles(layout, view))
	}
	return templates
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
