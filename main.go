package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/builds"
	"github.com/craigfurman/woodhouse-ci/db"
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"
	"github.com/craigfurman/woodhouse-ci/web"

	"github.com/codegangsta/negroni"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	must(err)
	distBase := filepath.Join(dir, "..")

	port := flag.Uint("port", 8080, "port to listen on")
	templateDir := flag.String("templateDir", filepath.Join(distBase, "web", "templates"), "path to html templates")
	storeDir := flag.String("storeDir", filepath.Join(distBase, "db"), "directory for saving persistent data")
	buildsDir := flag.String("buildsDir", filepath.Join(distBase, "builds"), "directory for saving build output")
	assetsDir := flag.String("assetsDir", filepath.Join(distBase, "web", "assets"), "path to static web assets")
	gooseCmd := flag.String("gooseCmd", filepath.Join(distBase, "bin", "goose"), `path to "goose" database migration tool`)
	flag.Parse()

	dbDir := filepath.Join(*storeDir, "sqlite")
	must(os.MkdirAll(dbDir, 0755))

	migrateCmd := exec.Command(*gooseCmd, "up")
	migrateCmd.Dir = filepath.Join(*storeDir, "..")
	must(migrateCmd.Run())

	jobRepo, err := db.NewJobRepository(filepath.Join(dbDir, "store.db"))
	must(err)

	handler := web.New(&jobs.Service{
		JobRepository: jobRepo,
		Runner:        runner.NewDockerRunner(),
		BuildRepository: &builds.Repository{
			BuildsDir: *buildsDir,
		},
	}, *templateDir)

	server := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.NewStatic(http.Dir(*assetsDir)))
	server.UseHandler(handler)
	server.Run(fmt.Sprintf("0.0.0.0:%d", *port))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
