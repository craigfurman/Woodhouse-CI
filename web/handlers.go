package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/craigfurman/woodhouse-ci/blockingio"
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	"github.com/gorilla/mux"
)

//go:generate counterfeiter -o fake_job_service/fake_job_service.go . JobService
type JobService interface {
	ListJobs() ([]jobs.Job, error)
	Save(job *jobs.Job) error
	RunJob(id string) error
	FindBuild(jobId string, buildNumber int) (jobs.Build, error)
	Stream(jobId string, buildNumber int, streamOffset int64) (*blockingio.BlockingReader, error)
}

type Handler struct {
	*mux.Router

	jobService JobService
	templates  map[string]*template.Template
}

func New(jobService JobService, templateDir string) *Handler {
	templates := parseTemplates(templateDir)
	router := mux.NewRouter()

	h := &Handler{
		Router:     router,
		templates:  templates,
		jobService: jobService,
	}

	h.HandleFunc("/", h.rootHandler).Methods("GET")
	h.HandleFunc("/jobs", h.listJobs).Methods("GET")
	h.HandleFunc("/jobs/new", h.newJob).Methods("GET")
	h.HandleFunc("/jobs", h.createJob).Methods("POST")
	h.HandleFunc("/jobs/{jobId}/builds/{buildId}", h.showBuild).Methods("GET")
	h.HandleFunc("/jobs/{jobId}/builds/{buildId}/output", h.streamBuild).Methods("GET")

	return h
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/jobs", 302)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	if list, err := h.jobService.ListJobs(); err == nil {
		h.renderTemplate("list_jobs", list, w)
	} else {
		h.renderErrPage("listing jobs", err, w, r)
	}
}

func (h *Handler) newJob(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate("new_job", nil, w)
}

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request) {
	job := jobs.Job{
		Name:          r.FormValue("name"),
		Command:       r.FormValue("command"),
		DockerImage:   r.FormValue("dockerImage"),
		GitRepository: r.FormValue("gitRepo"),
	}

	if err := h.jobService.Save(&job); err != nil {
		h.renderErrPage("saving job", err, w, r)
		return
	}

	if err := h.jobService.RunJob(job.ID); err == nil {
		http.Redirect(w, r, fmt.Sprintf("/jobs/%s/builds/1", job.ID), 302)
	} else {
		h.renderErrPage("running job", err, w, r)
	}
}

func (h *Handler) showBuild(w http.ResponseWriter, r *http.Request) {
	jobId := mux.Vars(r)["jobId"]
	buildIdStr := mux.Vars(r)["buildId"]
	buildId, err := strconv.Atoi(buildIdStr)
	must(err)
	if runningJob, err := h.jobService.FindBuild(jobId, buildId); err == nil {
		buildView := helpers.PresentableJob(runningJob)
		buildView.BuildNumber = buildIdStr
		buildView.BytesAlreadyReceived = len(runningJob.Output)
		h.renderTemplate("job_output", buildView, w)
	} else {
		h.renderErrPage("retrieving job", err, w, r)
	}
}

func (h *Handler) streamBuild(w http.ResponseWriter, r *http.Request) {
	jobId := mux.Vars(r)["jobId"]
	buildId, err := strconv.Atoi(mux.Vars(r)["buildId"])
	must(err)

	must(r.ParseForm())
	streamOffset, err := strconv.Atoi(r.Form.Get("offset"))
	must(err)

	streamer, err := h.jobService.Stream(jobId, buildId, int64(streamOffset))
	must(err)

	w.Header().Set("Content-Type", "text/event-stream\n\n")

	for {
		bytes, done := streamer.Next()
		_, err = w.Write([]byte(eventMessage("output", helpers.SanitisedHTML(string(bytes)))))
		must(err)

		w.(http.Flusher).Flush()

		if done {
			break
		}
	}

	must(streamer.Close())

	build, err := h.jobService.FindBuild(jobId, buildId)
	must(err)

	w.Write([]byte(eventMessage("end", helpers.Message(build))))
}

func eventMessage(eventName, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

type Error struct {
	Error string
}

func (h Handler) renderErrPage(message string, err error, w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: %s: %v", message, err)
	w.WriteHeader(500)
	h.renderTemplate("error", Error{Error: err.Error()}, w)
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
