package main

import (
	"log"
	"net/http"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}

func main() {
	http.HandleFunc("/", testHandler)
	log.Fatal(http.ListenAndServe(":8008", nil))
}
