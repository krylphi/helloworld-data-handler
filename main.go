package main

import (
	"github.com/krylphi/helloworld-data-handler/internal/stream"
	"github.com/krylphi/helloworld-data-handler/internal/stream/aws"
	"github.com/krylphi/helloworld-data-handler/internal/stream/azure"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/krylphi/helloworld-data-handler/internal/routing"
	"github.com/krylphi/helloworld-data-handler/internal/utils"

	"github.com/valyala/fasthttp"
)

func main() {
	storage := strings.ToLower(utils.GetEnvDef("STORAGE", "aws"))
	var handler stream.Handler
	switch storage {
	case "aws":
		handler = aws.NewStreamHandler()
	case "azure":
		errorsChan := make(chan error)
		azure.NewStreamHandler(errorsChan)
		go func() {
			for err := range errorsChan {
				log.Println(err)
			}
		}()
	default:
		handler = aws.NewStreamHandler()
	}
	router := routing.NewRouter(handler)
	addr := utils.Concat(utils.GetEnvDef("ADDR", "0.0.0.0"), ":", utils.GetEnvDef("PORT", "8902"))
	go func() {
		err := fasthttp.ListenAndServe(addr, router.HTTPRouter)
		if err != nil {
			log.Print(err.Error())
			log.Fatal("error handling")
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Server ...")

	shutdown(router)

	log.Println("Server exiting")

}

func shutdown(r routing.Router) {
	r.Shutdown()
}
