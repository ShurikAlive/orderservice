package main

import (
	"Test/transport"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

import log "github.com/sirupsen/logrus"

func main() {

	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile("my.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		log.SetOutput(file)
		defer file.Close()
	}

	db := transport.GetDBInstance()
	if db == nil {
		log.Info("Error DB connect")
		return
	} // Инициализируем БД

	a, err := parseEnv()
	serverUrl := a.ServeRESTAddress

	log.WithFields(log.Fields{"url": serverUrl}).Info("starting the server")

	killSignalChan := getKillSignalChan()
	srv := startServer(serverUrl)

	waitForKillSignal(killSignalChan)
	srv.Shutdown(context.Background())
}

func startServer(serverUrl string) *http.Server {
	router := transport.Router()
	log.Fatal(http.ListenAndServe(serverUrl, router))

	srv := &http.Server{Addr: serverUrl, Handler: router}
	go func() {
		log.Fatal(srv.ListenAndServe())
	} ()

	return srv
}

func getKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChan
}

func waitForKillSignal(killSignalChan <-chan os.Signal) {
	killSignal := <-killSignalChan
	switch killSignal{
	case os.Interrupt:
		log.Info("got SIGINT...")
	case syscall.SIGTERM:
		log.Info("got SIGTERM...")
	}
}
