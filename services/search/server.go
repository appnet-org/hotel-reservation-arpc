package search

import (
	// "encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	// "os"
	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	hotel "github.com/appnetorg/hotel-reservation-arpc/proto"

	"context"

	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
)

const _ = "srv-search"

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
	server, err := rpc.NewServer(s.IpAddr+":"+strconv.Itoa(s.Port), serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
		return err
	}

	hotel.RegisterSearchServer(server, s)

	// init arpc clients before starting the server
	if err := s.initGeoClient("geo.default.svc.cluster.local:11003"); err != nil {
		return err
	}
	if err := s.initRateClient("rate.default.svc.cluster.local:11004"); err != nil {
		return err
	}

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

func (s *Server) initGeoClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create geo aRPC client: %v", err)
	}

	s.geoClient = hotel.NewGeoClient(client)
	return nil
}

func (s *Server) initRateClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create rate aRPC client: %v", err)
	}

	s.rateClient = hotel.NewRateClient(client)
	return nil
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *hotel.SearchRequest) (*hotel.SearchResult, context.Context, error) {
	if s.geoClient == nil {
		log.Error().Msg("geo client not initialized")
		return nil, ctx, fmt.Errorf("geo client not initialized")
	}

	// Add timeout to context to prevent hanging
	geoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	nearby, err := s.geoClient.NearbyGeo(geoCtx, &hotel.NearbyRequest{
		Lat:       req.Lat,
		Lon:       req.Lon,
		Latstring: fmt.Sprintf("%f", req.Lat),
	})
	if err != nil {
		log.Error().Msgf("geoClient.NearbyGeo failed: %v", err)
		return nil, ctx, fmt.Errorf("geoClient.NearbyGeo failed: %w", err)
	}

	// find rates for hotels
	if s.rateClient == nil {
		log.Error().Msg("rate client not initialized")
		return nil, ctx, fmt.Errorf("rate client not initialized")
	}

	// Add timeout to context for rate client call
	rateCtx, cancelRate := context.WithTimeout(ctx, 10*time.Second)
	defer cancelRate()

	rates, err := s.rateClient.GetRates(rateCtx, &hotel.GetRatesRequest{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		log.Error().Msgf("rateClient.GetRates failed: %v", err)
		return nil, ctx, fmt.Errorf("rateClient.GetRates failed: %w", err)
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(hotel.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}

	return res, ctx, nil
}
