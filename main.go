package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/craigfurman/woodhouse-ci/db"
	"github.com/craigfurman/woodhouse-ci/web"

	"github.com/codegangsta/negroni"
)

func main() {
	port := flag.Uint("port", 0, "port to listen on")
	templateDir := flag.String("templateDir", "", "path to html templates")
	storeDir := flag.String("storeDir", "", "directory for saving persistent data")
	flag.Parse()

	migrateCmd := exec.Command("goose", "up")
	migrateCmd.Dir = filepath.Join(*storeDir, "..")
	must(migrateCmd.Run())

	dbDir := filepath.Join(*storeDir, "sqlite")
	must(os.MkdirAll(dbDir, 0755))

	jobRepo, err := db.NewJobRepository(filepath.Join(dbDir, "store.db"))
	must(err)

	handler := web.New(jobRepo, *templateDir)

	server := negroni.Classic()
	server.UseHandler(handler)
	server.Run(fmt.Sprintf("0.0.0.0:%d", *port))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
