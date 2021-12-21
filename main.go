package main

import (
	"log"
	"os"

	"github.com/tsuru/tsuru-prometheus-api/api"
	"github.com/tsuru/tsuru-prometheus-api/service"
)

func main() {
	tsuruHost := os.Getenv("TSURU_HOST")
	tsuruToken := os.Getenv("TSURU_TOKEN")

	authUser := os.Getenv("AUTH_USER")
	if authUser == "" {
		authUser = "admin"
	}

	authPassword := os.Getenv("AUTH_PASSWORD")
	if authPassword == "" {
		authPassword = "admin"
	}

	if tsuruHost == "" || tsuruToken == "" {
		log.Fatalln("TSURU_HOST and TSURU_TOKEN must be set")
	}

	svc := service.NewService(tsuruHost, tsuruToken)

	server := api.NewServer(api.ServerOpts{
		Service:      svc,
		AuthUser:     authUser,
		AuthPassword: authPassword,
	})
	server.Run()
}
