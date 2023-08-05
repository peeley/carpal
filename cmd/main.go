package main

import (
	"log"
	"net/http"
	"os"

	"github.com/peeley/carpal/internal/config"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}

func main() {
	configWizard := config.NewConfigWizard(os.Getenv("CONFIG_FILE"))
	config, err := configWizard.GetConfiguration()
	if err != nil {
		log.Fatalf("could not load configuration: %v", err)
	}

	log.Println(config)

	http.HandleFunc("/", testHandler)
	log.Fatal(http.ListenAndServe(":8008", nil))
}
