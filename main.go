package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"prova/app"
	"prova/http"
	"prova/postgres"

	"github.com/inconshreveable/log15"
)

func main() {

	ctx, cancel := context.WithCancel(app.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	postgresDB := postgres.NewDB(os.Getenv("POSTGRES_URL"))

	if err := postgresDB.Open(); err!= nil{
		panic(err)
	}
	defer postgresDB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	postgresUserService := postgres.NewUserService(postgresDB)

	logger := log15.New()

	server := http.NewServerAPI()

	server.Addr = fmt.Sprintf(":%s", port)
	server.BaseURL = os.Getenv("BASE_URL")
	server.LogService = logger.New("module", "http")
	server.UserService = postgresUserService

	if err := server.Open(); err != nil {
		panic(err)
	}

	logger.Info("Starting server", "port", port)

	<-ctx.Done()

	if err := server.Close(); err != nil {
		panic(err)
	}

	logger.Info("Closing server")

}
