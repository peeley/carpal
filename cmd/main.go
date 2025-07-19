package main

import (
	"log"
	"net/http"
	"os"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/driver/file"
	"github.com/peeley/carpal/internal/driver/ldap"
	"github.com/peeley/carpal/internal/driver/sql"
	"github.com/peeley/carpal/internal/handler"
)

func main() {
	fileLocation := os.Getenv("CONFIG_FILE")
	if fileLocation == "" {
		fileLocation = "/etc/carpal/config.yml"
	}

	configWizard := config.NewConfigWizard(fileLocation)
	config, err := configWizard.GetConfiguration()
	if err != nil {
		log.Fatalf("could not load configuration: %v", err)
	}

	var driver driver.Driver
	switch config.Driver {
	case "file":
		driver = file.NewFileDriver(*config)
	case "ldap":
		driver = ldap.NewLDAPDriver(*config)
	case "sql":
		var err error
		driver, err = sql.NewSQLDriver(*config)
		if err != nil {
			log.Fatalf("failed to initialize SQL driver: %v", err)
		}
	default:
		log.Fatalf("driver `%s` is invalid", config.Driver)
	}

	handler := handler.NewResourceHandler(driver)
	http.HandleFunc("/", handler.Handle)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8008"
	}

	log.Printf("launching carpal server on port %v...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
