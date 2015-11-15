package main

import (
	"fmt"
	"net/http"
)

func main() {
	port := "3000"
	http.Handle("/", http.FileServer(http.Dir("./html")))
	fmt.Println("Listening at http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}
