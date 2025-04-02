package main

import (
	"log"

	"github.com/grigory222/scraptor/src/config"
	"github.com/grigory222/scraptor/src/handlers"
	"github.com/grigory222/scraptor/src/repository"
	"github.com/grigory222/scraptor/src/service"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.Load()

	db := repository.NewPostgres(cfg.DB)

	svc := service.NewService(db)

	e := echo.New()
	handlers.RegisterRoutes(e, svc)

	log.Fatal(e.Start(cfg.ServerAddr))
}
