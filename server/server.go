package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/Reskill-2022/volunteering/config"
	"github.com/Reskill-2022/volunteering/controllers"
	"github.com/Reskill-2022/volunteering/linkedin"
	"github.com/Reskill-2022/volunteering/repository"
)

func registerRoutes(e *echo.Echo, cts *controllers.Container, rc *repository.Container, service linkedin.Service) {
	e.Use(middleware.Logger())
	// allow all origins
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	api := e.Group("/volunteering")

	api.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Backend! OK")
	})
	{
		users := api.Group("/users")

		users.POST("", cts.UserController.CreateUser(rc.UserRepository, service))
		users.PUT("/:email", cts.UserController.UpdateUser(rc.UserRepository, rc.UserRepository))
		users.GET("/:email", cts.UserController.GetUser(rc.UserRepository))
	}
}

func Start(logger zerolog.Logger, env config.Environment, cts *controllers.Container, rc *repository.Container, service linkedin.Service) error {
	e := echo.New()

	registerRoutes(e, cts, rc, service)

	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Addr:         fmt.Sprintf(":%s", env[config.Port]),
	}

	// gracefully start server
	go func() {
		if err := e.StartServer(srv); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// listen for ctrl+c and gracefully shutdown server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Failed to shutdown server")
	}

	logger.Info().Msg("Server gracefully shutdown")
	return nil
}
