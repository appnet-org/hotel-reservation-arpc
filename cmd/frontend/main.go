package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/appnet-org/arpc-quic/pkg/logging"
	"github.com/appnetorg/hotel-reservation-arpc/services/frontend"
	"github.com/appnetorg/hotel-reservation-arpc/tracing"
	"github.com/appnetorg/hotel-reservation-arpc/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// getLoggingConfig reads logging configuration from environment variables with defaults
func getLoggingConfig() *logging.Config {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "console"
	}

	return &logging.Config{
		Level:  level,
		Format: format,
	}
}

func main() {
	tune.Init()
	err := logging.Init(getLoggingConfig())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logging: %v", err))
	}
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Caller().Logger()

	log.Info().Msg("Reading config...")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Error().Msgf("Got error while reading config: %v", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	serv_port, _ := strconv.Atoi(result["FrontendPort"])
	serv_ip := result["FrontendIP"]
	knative_dns := result["KnativeDomainName"]

	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])
	var (
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
	)
	flag.Parse()
	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "frontend", *jaegeraddr)
	tracer, err := tracing.Init("frontend", *jaegeraddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	srv := &frontend.Server{
		KnativeDns: knative_dns,
		Tracer:     tracer,
		IpAddr:     serv_ip,
		Port:       serv_port,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
