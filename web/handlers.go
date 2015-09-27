package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/craigfurman/woodhouse-ci/chunkedio"
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	"github.com/gorilla/mux"
)

//go:generate counterfeiter -o fake_job_service/fake_job_service.go . JobService
type JobService interface {
	AllLatestBuilds() ([]jobs.Build, error)
	Save(job *jobs.Job) error
	RunJob(id string) (int, error)
	FindBuild(jobId string, buildNumber int) (jobs.Build, error)
	HighestBuild(jobId string) (int, error)
	Stream(jobId string, buildNumber int, streamOffset int64) (*chunkedio.ChunkedReader, error)
}

type Handler struct {
	*mux.Router

	jobService   JobService
	templates    map[string]*template.Template
	templateSets map[string][]string
}

func New(jobService JobService, templateDir string, preloadTemplates bool) *Handler {
	templateSets := collectTemplates(templateDir)
	templates := make(map[string]*template.Template)
	if preloadTemplates {
		for viewName, templateSet := range templateSets {
			templates[viewName] = template.Must(template.ParseFiles(templateSet...))
		}
	}

	router := mux.NewRouter()

	h := &Handler{
		Router:       router,
		templates:    templates,
		templateSets: templateSets,
		jobService:   jobService,
	}

	h.HandleFunc("/", h.rootHandler).Methods("GET")
	h.HandleFunc("/jobs", h.listJobs).Methods("GET")
	h.HandleFunc("/jobs/status", h.listJobStatuses).Methods("GET")
	h.HandleFunc("/jobs/new", h.newJob).Methods("GET")
	h.HandleFunc("/jobs", h.createJob).Methods("POST")
	h.HandleFunc("/jobs/{jobId}/builds", h.createBuild).Methods("POST")
	h.HandleFunc("/jobs/{jobId}/builds/{buildId}", h.showBuild).Methods("GET")
	h.HandleFunc("/jobs/{jobId}/builds/{buildId}/output", h.streamBuild).Methods("GET")

	return h
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/jobs", 302)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	if list, err := h.jobService.AllLatestBuilds(); err == nil {
		type cell struct {
			ID     string
			Name   string
			Status string
		}

		cells := [][]cell{}
		for _, buildRow := range helpers.JobGrid(list) {
			cellRow := []cell{}
			for _, build := range buildRow {
				cellRow = append(cellRow, cell{
					ID:     build.ID,
					Name:   build.Name,
					Status: helpers.Classes(build),
				})
			}
			cells = append(cells, cellRow)
		}

		p := struct {
			BuildRows [][]cell
		}{
			BuildRows: cells,
		}
		h.renderTemplate("list_jobs", p, w)
	} else {
		h.renderErrPage("listing jobs", err, w, r)
	}
}

func (h *Handler) listJobStatuses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream\n\n")

	for {
		list, err := h.jobService.AllLatestBuilds()
		must(err)

		statuses := make(map[string]string)
		for _, build := range list {
			statuses[build.ID] = helpers.Classes(build)
			msg, err := json.Marshal(statuses)
			must(err)
			if _, err := w.Write([]byte(eventMessage("jobs", string(msg)))); err != nil {
				log.Printf("trying to write job statuses JSON. assuming remote end hung up. Cause: %v\n", err)
				return
			}

			w.(http.Flusher).Flush()
			time.Sleep(time.Millisecond * 100)
		}
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

	if buildNumber, err := h.jobService.RunJob(job.ID); err == nil {
		http.Redirect(w, r, fmt.Sprintf("/jobs/%s/builds/%d", job.ID, buildNumber), 302)
	} else {
		h.renderErrPage("running job", err, w, r)
	}
}

func (h *Handler) createBuild(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["jobId"]
	if buildNumber, err := h.jobService.RunJob(jobID); err == nil {
		http.Redirect(w, r, fmt.Sprintf("/jobs/%s/builds/%d", jobID, buildNumber), 302)
	} else {
		h.renderErrPage("running job", err, w, r)
	}
}

func (h *Handler) showBuild(w http.ResponseWriter, r *http.Request) {
	jobId := mux.Vars(r)["jobId"]
	buildIdStr := mux.Vars(r)["buildId"]

	if buildIdStr == "latest" {
		buildNumber, err := h.jobService.HighestBuild(jobId)
		must(err)
		http.Redirect(w, r, fmt.Sprintf("/jobs/%s/builds/%d", jobId, buildNumber), 302)
		return
	}

	buildId, err := strconv.Atoi(buildIdStr)
	must(err)

	if build, err := h.jobService.FindBuild(jobId, buildId); err == nil {
		sanitizedOutput := helpers.SanitisedHTML(build.Output)
		highestBuildNumber, err := h.jobService.HighestBuild(jobId)
		if err != nil {
			h.renderErrPage("finding highest build number for job", err, w, r)
			return
		}
		buildNumbers := []int{}
		for i := highestBuildNumber; i > 0; i-- {
			buildNumbers = append(buildNumbers, i)
		}

		buildView := struct {
			Build                jobs.Build
			BuildNumber          int
			Output               template.HTML
			BytesAlreadyReceived int
			ExitMessage          string
			BuildNumbers         []int
		}{
			Build:                build,
			BuildNumber:          buildId,
			Output:               sanitizedOutput,
			BytesAlreadyReceived: len(sanitizedOutput),
			ExitMessage:          helpers.Message(build),
			BuildNumbers:         buildNumbers,
		}
		h.renderTemplate("show_build", buildView, w)
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
		_, err = w.Write([]byte(eventMessage("output", string(helpers.SanitisedHTML(bytes)))))
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
	if t, ok := h.templates[name]; ok {
		must(t.Execute(w, pageObject))
	} else {
		log.Printf("template %s not found, parsing. If debug mode is not active, then something has gone wrong!\n", name)
		must(template.Must(template.ParseFiles(h.templateSets[name]...)).Execute(w, pageObject))
	}
}

func collectTemplates(templateDir string) map[string][]string {
	templateFor := func(viewType, view string) string {
		return filepath.Join(templateDir, viewType, view+".html")
	}

	layoutFor := func(layout string) string {
		return templateFor("layouts", layout)
	}

	viewFor := func(view string) string {
		return templateFor("views", view)
	}

	listJobs := "list_jobs"
	newJob := "new_job"
	showBuild := "show_build"
	errorPage := "error"

	return map[string][]string{
		listJobs:  {layoutFor("outer"), viewFor(listJobs)},
		newJob:    {layoutFor("outer"), layoutFor("single_column"), viewFor(newJob)},
		showBuild: {layoutFor("outer"), layoutFor("single_column"), viewFor(showBuild)},
		errorPage: {layoutFor("outer"), layoutFor("single_column"), viewFor(errorPage)},
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
