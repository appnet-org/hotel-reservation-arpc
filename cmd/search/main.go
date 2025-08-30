package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"strconv"

	"github.com/appnetorg/hotel-reservation-arpc/services/search"
	"github.com/appnetorg/hotel-reservation-arpc/tracing"
	"github.com/appnetorg/hotel-reservation-arpc/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	tune.Init()
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

	serv_port, _ := strconv.Atoi(result["SearchPort"])
	serv_ip := result["SearchIP"]
	knative_dns := result["KnativeDomainName"]
	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])

	var (
		// port       = flag.Int("port", 8082, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "search", *jaegeraddr)
	tracer, err := tracing.Init("search", *jaegeraddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	srv := &search.Server{
		Tracer: tracer,
		// Port:     *port,
		Port:       serv_port,
		IpAddr:     serv_ip,
		KnativeDns: knative_dns,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
