package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"

	"github.com/gorilla/mux"
)

//go:generate counterfeiter -o fake_job_service/fake_job_service.go . JobService
type JobService interface {
	RunJob(id string) (jobs.RunningJob, error)
	Save(job *jobs.Job) error
}

type Handler struct {
	*mux.Router

	templates map[string]*template.Template
}

func New(jobService JobService, templateDir string) *Handler {
	templates := parseTemplates(templateDir)
	router := mux.NewRouter()

	handler := &Handler{
		Router:    router,
		templates: templates,
	}

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("list_jobs", nil, w)
	}).Methods("GET")

	router.HandleFunc("/jobs/new", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("create_job", nil, w)
	}).Methods("GET")

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		job := jobs.Job{
			Name:    r.FormValue("name"),
			Command: r.FormValue("command"),
		}
		if err := jobService.Save(&job); err == nil {
			http.Redirect(w, r, fmt.Sprintf("/jobs/%s/output", job.ID), 302)
		} else {
			errPage("saving job", err, w, r)
		}
	}).Methods("POST")

	router.HandleFunc("/jobs/{id}/output", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if completedJob, err := jobService.RunJob(id); err == nil {
			handler.renderTemplate("job_output", completedJob, w)
		} else {
			errPage("retrieving job", err, w, r)
		}
	}).Methods("GET")

	router.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		handler.renderTemplate("error", nil, w)
	}).Methods("GET")

	return handler
}

func errPage(message string, err error, w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: %s: %v", message, err)
	http.Redirect(w, r, "/error", 302)
}

func (h Handler) renderTemplate(name string, pageObject interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	must(h.templates[name].Execute(w, pageObject))
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
