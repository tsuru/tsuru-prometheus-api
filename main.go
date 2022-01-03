package main

import (
	"log"
	"os"

	"github.com/tsuru/tsuru-prometheus-api/api"
	"github.com/tsuru/tsuru-prometheus-api/service"
)

func main() {
	kubernetesToken := os.Getenv("KUBERNETES_TOKEN")

	tsuruTarget := os.Getenv("TSURU_TARGET")
	tsuruToken := os.Getenv("TSURU_TOKEN")

	authUser := os.Getenv("AUTH_USER")
	if authUser == "" {
		authUser = "admin"
	}

	authPassword := os.Getenv("AUTH_PASSWORD")
	if authPassword == "" {
		authPassword = "admin"
	}

	if tsuruTarget == "" || tsuruToken == "" {
		log.Fatalln("TSURU_TARGET and TSURU_TOKEN must be set")
	}

	svc := service.NewService(tsuruToken, kubernetesToken)

	server := api.NewServer(api.ServerOpts{
		Service:      svc,
		AuthUser:     authUser,
		AuthPassword: authPassword,
	})
	server.Run()
}
