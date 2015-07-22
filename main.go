package main

import (
	"flag"
	"fmt"

	"github.com/codegangsta/negroni"
	"github.com/craigfurman/woodhouse-ci/web"
)

func main() {
	port := flag.Uint("port", 0, "port to listen on")
	templateDir := flag.String("templateDir", "", "path to html templates")
	flag.Parse()

	handler := web.New(*templateDir)

	server := negroni.Classic()
	server.UseHandler(handler)
	server.Run(fmt.Sprintf("0.0.0.0:%d", *port))
}
