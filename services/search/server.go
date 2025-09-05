package search

import (
	// "encoding/json"
	"fmt"
	// F"io/ioutil"

	"github.com/rs/zerolog/log"

	// "os"
	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	hotel "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"

	"context"

	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
)

const name = "srv-search"

// Server implments the search service
type Server struct {
	rateClient hotel.RateClient
	geoClient  hotel.GeoClient

	Tracer     opentracing.Tracer
	Port       int
	IpAddr     string
	KnativeDns string
	uuid       string
}

// mustEmbedUnimplementedSearchServer is a placeholder method to satisfy the SearchServer interface.

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer(s.IpAddr, serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
	}

	hotel.RegisterSearchServer(server, &Server{})

	server.Start()

	// init arpc clients
	if err := s.initGeoClient("geo.default.svc.cluster.local:8083"); err != nil {
		return err
	}
	if err := s.initRateClient("rate.default.svc.cluster.local:8084"); err != nil {
		return err
	}

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

func (s *Server) initGeoClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClient(serializer, name, nil)
	if err != nil {
		return fmt.Errorf("failed to create aRPC client: %v", err)
	}

	s.geoClient = hotel.NewGeoClient(client)
	return nil
}

func (s *Server) initRateClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClient(serializer, name, nil)
	if err != nil {
		return fmt.Errorf("failed to create aRPC client: %v", err)
	}

	s.rateClient = hotel.NewRateClient(client)
	return nil
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *hotel.SearchRequest) (*hotel.SearchResult, context.Context, error) {
	// find nearby hotels
	log.Trace().Msg("Nearby got a message")

	log.Trace().Msgf("nearby lat = %f", req.Lat)
	log.Trace().Msgf("nearby lon = %f", req.Lon)

	nearby, err := s.geoClient.Nearby(ctx, &hotel.NearbyRequest{
		Lat:       req.Lat,
		Lon:       req.Lon,
		Latstring: fmt.Sprintf("%f", req.Lat),
	})
	if err != nil {
		return nil, ctx, err
	}

	for _, hid := range nearby.HotelIds {
		log.Trace().Msgf("get Nearby hotelId = %s", hid)
	}

	// find rates for hotels
	rates, err := s.rateClient.GetRates(ctx, &hotel.GetRatesRequest{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		return nil, ctx, err
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(hotel.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		// log.Trace().Msgf("gået RatePlan HotelId = %s, Code = %s", ratePlan.HotelId, ratePlan.Code)
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}

	return res, ctx, nil
}
