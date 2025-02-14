package main

import (
	"context"
	"errors"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"os"
	"strings"

	"github.com/pajtaand/dmap-zero/internal/common/constants"
	"github.com/pajtaand/dmap-zero/internal/controller/app"
)

const (
	exitTimeout = 5 * time.Second

	defaultCertFilePath = "/certs/ca-cert.pem"
	defaultKeyFilePath  = "/certs/ca-key.pem"
)

var pairListRegex = regexp.MustCompile(`^([^\s:]+:[^\s:]+)(,[^\s:]+:[^\s:]+)*$`)

func main() {
	zerolog.DefaultContextLogger = &log.Logger

	// Load values from environment variables using constants
	apiCertFile := os.Getenv(constants.ControllerEnvAPICertFile)
	if apiCertFile == "" {
		apiCertFile = defaultCertFilePath
	}

	apiKeyFile := os.Getenv(constants.ControllerEnvAPIKeyFile)
	if apiKeyFile == "" {
		apiKeyFile = defaultKeyFilePath
	}

	enrollmentToken := os.Getenv(constants.ControllerEnvEnrollmentToken)
	apiCredentials := os.Getenv(constants.ControllerEnvAPICredentials)

	// Check required fields
	if enrollmentToken == "" {
		log.Error().Msgf("Error: %s environment variable is required", constants.ControllerEnvEnrollmentToken)
		os.Exit(1)
	}

	if apiCredentials == "" {
		log.Error().Msgf("Error: %s environment variable is required", constants.ControllerEnvAPICredentials)
		os.Exit(1)
	}

	credentials, err := parseAPICredentials(apiCredentials)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't parse credentials.")
		os.Exit(1)
	}

	// Set up the configuration
	cfg := &app.ControllerAppConfig{}
	cfg.ApiCredentials = credentials
	cfg.OpenZiti.KeyAlg = "RSA"
	cfg.OpenZiti.EnrollmentToken = enrollmentToken
	cfg.RESTapi.Address = constants.ControllerAPIAddress
	cfg.RESTapi.CertFile = apiCertFile
	cfg.RESTapi.KeyFile = apiKeyFile
	cfg.MetricsApi.Address = constants.ControllerMetricsAPIAddress
	cfg.MetricsApi.CertFile = apiCertFile
	cfg.MetricsApi.KeyFile = apiKeyFile

	controllerApp, err := app.NewControllerApp(cfg)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := controllerApp.Setup(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := controllerApp.Run(ctx); err != nil {
			panic(err)
		}
	}()

	// register signal to exit gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		wg.Add(1)
		defer wg.Done()

		log.Info().Msgf("Signal recevied. Shutting down with timeout=%v", exitTimeout)
		ctx, cancel := context.WithTimeout(ctx, exitTimeout)
		defer cancel()
		go func() {
			<-ctx.Done()
			if ctx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := controllerApp.Stop(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	log.Info().Msg("Cleaning up...")
	if err := controllerApp.Clean(ctx); err != nil {
		panic(err)
	}
}

func parseAPICredentials(credentials string) (map[string]string, error) {
	if credentials != "" {
		if !pairListRegex.MatchString(credentials) {
			return nil, errors.New("invalid credentials format, expected format is username:password[,username:password...]")
		}
	}

	tuples := strings.Split(credentials, ",")
	parsedCredentials := map[string]string{}

	for _, tuple := range tuples {
		parts := strings.Split(tuple, ":")
		if len(parts) == 2 {
			parsedCredentials[parts[0]] = parts[1]
		}
	}

	return parsedCredentials, nil
}
