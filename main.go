package main

import (
	"github.com/krylphi/helloworld-data-handler/internal/routing"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	router := routing.NewRouter()
	go func() {
		err := fasthttp.ListenAndServe(":8902", router.HttpRouter)
		if err != nil {
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

	router.Shutdown()

	log.Println("Server exiting")

}

func shutdown(r routing.Router) {
	r.Shutdown()
}
