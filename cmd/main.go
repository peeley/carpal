package main

import (
	"log"
	"net/http"
	"os"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/driver/file"
	"github.com/peeley/carpal/internal/handler"
)

func main() {
	configWizard := config.NewConfigWizard(os.Getenv("CONFIG_FILE"))
	config, err := configWizard.GetConfiguration()
	if err != nil {
		log.Fatalf("could not load configuration: %v", err)
	}

	var driver driver.Driver
	switch config.Driver {
	case "file":
		driver = file.NewFileDriver(*config)
	default:
		log.Fatalf("driver `%s` is invalid", config.Driver)
	}

	handler := handler.NewResourceHandler(driver)
	http.HandleFunc("/", handler.Handle)

	log.Fatal(http.ListenAndServe(":8008", nil))
}
