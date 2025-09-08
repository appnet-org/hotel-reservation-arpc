package geo

import (
	// "encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"context"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	"github.com/google/uuid"
	"github.com/hailocab/go-geoindex"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 5
)

// Server implements the geo service
type Server struct {
	index *geoindex.ClusteringIndex
	uuid  string

	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.index == nil {
		s.index = newGeoIndex(s.MongoSession)
	}

	s.uuid = uuid.New().String()

	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer(s.IpAddr+":"+strconv.Itoa(s.Port), serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterGeoServer(server, s)

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// Nearby returns all hotels within a given distance.
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.NearbyResult, context.Context, error) {
	log.Info().Msgf("In geo getNearbyPoints, lat = %f, lon = %f, latstring = %s", req.Lat, req.Lon, req.Latstring)

	var (
		points = s.getNearbyPoints(ctx, float64(req.Lat), float64(req.Lon))
		res    = &pb.NearbyResult{}
	)

	log.Trace().Msgf("geo after getNearbyPoints, len = %d", len(points))

	for _, p := range points {
		log.Trace().Msgf("In geo Nearby return hotelId = %s", p.Id())
		res.HotelIds = append(res.HotelIds, p.Id())
	}

	return res, ctx, nil
}

func (s *Server) getNearbyPoints(ctx context.Context, lat, lon float64) []geoindex.Point {

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.index.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

// newGeoIndex returns a geo index with points loaded
func newGeoIndex(session *mgo.Session) *geoindex.ClusteringIndex {
	// session, err := mgo.Dial("mongodb-geo")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	log.Trace().Msg("new geo newGeoIndex")

	s := session.Copy()
	defer s.Close()
	c := s.DB("geo-db").C("geo")

	var points []*point
	err := c.Find(bson.M{}).All(&points)
	if err != nil {
		log.Error().Msgf("Failed get geo data: %v", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

// Implement Point interface
func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }
