package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/builds"
	"github.com/craigfurman/woodhouse-ci/db"
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"
	"github.com/craigfurman/woodhouse-ci/vcs"
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

	bootMsg := ` _    _                 _ _                                 _____ _____
| |  | |               | | |                               /  __ \_   _|
| |  | | ___   ___   __| | |__   ___  _   _ ___  ___ ______| /  \/ | |
| |/\| |/ _ \ / _ \ / _` + "`" + ` | '_ \ / _ \| | | / __|/ _ \______| |     | |
\  /\  / (_) | (_) | (_| | | | | (_) | |_| \__ \  __/      | \__/\_| |_
 \/  \/ \___/ \___/ \__,_|_| |_|\___/ \__,_|___/\___|       \____/\___/
`

	fmt.Println(bootMsg)

	dbDir := filepath.Join(*storeDir, "sqlite")
	must(os.MkdirAll(dbDir, 0755))

	migrateCmd := exec.Command(*gooseCmd, "up")
	migrateCmd.Dir = filepath.Join(*storeDir, "..")
	must(migrateCmd.Run())

	jobRepo, err := db.NewJobRepository(filepath.Join(dbDir, "store.db"))
	must(err)

	// Only Interrupt handled, as this is available on all major platforms and is the most common way of stopping Woodhouse-CI
	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt)
	go func(c <-chan os.Signal) {
		log.Printf("Caught signal %s. Closing database connections. Goodbye!\n", <-c)
		must(jobRepo.Close())
		os.Exit(0)
	}(exitChan)

	handler := web.New(&jobs.Service{
		JobRepository: jobRepo,
		Runner:        runner.NewDockerRunner(vcs.GitCloner{}),
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
