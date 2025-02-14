package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pajtaand/dmap-zero/internal/agent/app"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultKeyAlg = "RSA"
	exitTimeout   = 5 * time.Second
)

func main() {
	zerolog.DefaultContextLogger = &log.Logger

	keyAlg := flag.String("key-alg", defaultKeyAlg, "Key algorithm for private keys generation")
	enrollmentToken := flag.String("jwt", "", "Enrollment token (JWT) (required)")

	flag.Parse()

	if flag.Arg(0) == "-h" || flag.Arg(0) == "--help" {
		flag.Usage()
		return
	}

	if *enrollmentToken == "" {
		fmt.Println("Error: enrollment token is required")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	agentApp, err := app.NewAgentApp(ctx, app.AgentAppConfig{
		KeyAlg: *keyAlg,
		JWT:    *enrollmentToken,
	})
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := agentApp.Run(ctx); err != nil {
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

		if err := agentApp.Stop(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	log.Info().Msg("Cleaning up...")
	if err := agentApp.Clean(ctx); err != nil {
		panic(err)
	}
}
