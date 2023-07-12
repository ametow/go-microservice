package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"yandex-team.ru/bstask/internal/infrastructure"
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatalf("Usage: %v config_filename\n", os.Args[0])
	}

	if err := initConfig(os.Args[1]); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	app := infrastructure.Setup()

	go func() {
		if err := app.Start(viper.GetString("port")); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("failed to listen: %s", err.Error())
		}
	}()

	log.Println("Application started!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Gracefully shutting down...")
	if err := app.Shutdown(ctx); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}
}

func initConfig(filename string) error {
	viper.AddConfigPath("config")
	viper.SetConfigName(filename)
	return viper.ReadInConfig()
}
