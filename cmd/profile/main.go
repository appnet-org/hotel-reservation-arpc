package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/appnet-org/arpc-quic/pkg/logging"
	"github.com/appnetorg/hotel-reservation-arpc/services/profile"
	"github.com/appnetorg/hotel-reservation-arpc/tracing"
	"github.com/appnetorg/hotel-reservation-arpc/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"time"
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

	log.Info().Msgf("Read database URL: %v", result["ProfileMongoAddress"])
	log.Info().Msg("Initializing DB connection...")
	mongo_session := initializeDatabase(result["ProfileMongoAddress"])
	defer mongo_session.Close()
	log.Info().Msg("Successfull")

	log.Info().Msgf("Read profile memcashed address: %v", result["ProfileMemcAddress"])
	log.Info().Msg("Initializing Memcashed client...")
	memc_client := tune.NewMemCClient2(result["ProfileMemcAddress"])
	log.Info().Msg("Successfull")

	serv_port, _ := strconv.Atoi(result["ProfilePort"])
	serv_ip := result["ProfileIP"]
	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])

	var (
		// port       = flag.Int("port", 8081, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger server addr")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "profile", *jaegeraddr)
	tracer, err := tracing.Init("profile", *jaegeraddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	srv := profile.Server{
		Tracer: tracer,
		// Port:     *port,
		Port:         serv_port,
		IpAddr:       serv_ip,
		MongoSession: mongo_session,
		MemcClient:   memc_client,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
