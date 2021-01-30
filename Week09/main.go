package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"Go-001/Week09/echo"
)

var mode = flag.String("mode", "server", "server or client")

func main() {
	flag.Parse()
	switch *mode {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		fmt.Printf("err mode")
	}
}

func runServer() {
	s := echo.NewServer(":8000")
	go func() {
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	sign := <-c
	log.Printf("app get a signal %s\n", sign.String())
	if err := s.Shutdown(); err != nil {
		log.Printf("server stop error(%v)\n", err)
	}
}

func runClient() {
	client := echo.NewClient(":8000", os.Stdin)
	go func() {
		if err := client.Connect(); err != nil {
			log.Println("client stop")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sign := <-c:
		log.Printf("app get a signal %s\n", sign.String())
		client.Shutdown()
	case <-client.Done:
	}
	log.Println("client stop")
}
