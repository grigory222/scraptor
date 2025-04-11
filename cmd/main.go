package main

import (
	"github.com/grigory222/scraptor/internal/config"
	"github.com/grigory222/scraptor/internal/http-server/handlers"
	"github.com/grigory222/scraptor/internal/repository"
	"github.com/grigory222/scraptor/internal/service"
	"github.com/labstack/echo/v4"

	slogpretty "github.com/grigory222/scraptor/internal/logger"
)

func main() {
	cfg := config.Load()

	log := slogpretty.NewLogger()

	db := repository.NewPostgres(cfg.DB, log)
	svc := service.NewService(db, log)

	e := echo.New()

	handlers.RegisterMiddlewares(e)
	handlers.RegisterRoutes(e, svc)

	e.Start(cfg.ServerAddr)
}
