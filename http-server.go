package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {

	port := flag.String("port", "80", "listening port")
	folder := flag.String("dir", "./html", "directory to serve")
	flag.Parse()

	if *port == "" {
		*port = "80"
	}

	if *folder == "" {
		*port = "./html"
	}

	http.Handle("/", http.FileServer(http.Dir(*folder)))
	fmt.Println("Listening at http://localhost:" + *port + " and serving directory: " + *folder)
	http.ListenAndServe(":"+(*port), nil)
}
