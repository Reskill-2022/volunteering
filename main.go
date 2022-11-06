package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/Reskill-2022/volunteering/config"
	"github.com/Reskill-2022/volunteering/controllers"
	"github.com/Reskill-2022/volunteering/linkedin"
	"github.com/Reskill-2022/volunteering/repository"
	"github.com/Reskill-2022/volunteering/server"
)

var defaultWriter = zerolog.ConsoleWriter{Out: os.Stdout}

func main() {
	appLogger := zerolog.New(defaultWriter).With().Timestamp().Logger()

	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	env, err := config.New()
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to load configs")
	}

	writeSAs(appLogger, env)

	cts := controllers.NewContainer(appLogger)
	rc := repository.NewContainer(appLogger)
	service := linkedin.New(appLogger, env)

	if err := server.Start(appLogger, env, cts, rc, service); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func writeSAs(appLogger zerolog.Logger, env config.Environment) {
	sa1, ok := env[config.ServiceAccount1]
	if !ok {
		appLogger.Fatal().Msg("Service account 1 not found")
	}

	sa2, ok := env[config.ServiceAccount2]
	if !ok {
		appLogger.Fatal().Msg("Service account 2 not found")
	}

	sa1OutFile := "service-account-1.json"
	sa2OutFile := "service-account-2.json"

	if err := os.WriteFile(sa1OutFile, []byte(sa1), 0644); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to write service account 1")
	}

	if err := os.WriteFile(sa2OutFile, []byte(sa2), 0644); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to write service account 2")
	}
}
