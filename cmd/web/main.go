package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ryzenlo/to2cloud/configs"
	app "ryzenlo/to2cloud/internal/http"
	"ryzenlo/to2cloud/internal/models"
	xlog "ryzenlo/to2cloud/internal/pkg/log"
)

func main() {
	configs.LoadConfigFile("./configs/config.yaml")
	//
	baseCtx := context.Background()
	secondCtx, cancelFunc := context.WithCancel(baseCtx)
	xlog.InitLogger(secondCtx, configs.Conf)
	//
	if err := models.InitDBClient(secondCtx, configs.Conf); err != nil {
		xlog.Logger.Errorln("about to quit:")
		xlog.Logger.Errorln("%v", err)
		cancelFunc()
		return
	}
	//
	xlog.Logger.Infoln("Starting server...")
	//
	r := app.GetHttpRouter()
	srv := http.Server{
		Handler: r,
		Addr:    ":9000",
	}
	//
	eChan := make(chan os.Signal, 1)
	signal.Notify(eChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	// wait for signal
	<-eChan
	xlog.Logger.Infoln("Closing log file...")
	xlog.Logger.Infoln("Shutdowning server...")
	cancelFunc()
	//
	c, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := srv.Shutdown(c); err != nil {
		log.Fatal(err)
	}
}
