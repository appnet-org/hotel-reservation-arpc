package recommendation

import (
	// "encoding/json"
	"fmt"

	pb "github.com/appnetorg/hotel-reservation-arpc/services/recommendation/proto"
	"github.com/google/uuid"
	"github.com/hailocab/go-geoindex"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "io/ioutil"
	"math"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	// "os"
	// "strings"
)

const name = "srv-recommendation"

// Server implements the recommendation service
type Server struct {
	hotels       map[string]Hotel
	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	uuid         string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.hotels == nil {
		s.hotels = loadRecommendations(s.MongoSession)
	}

	s.uuid = uuid.New().String()

	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer(s.IpAddr, serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterRecommendationServer(server, &Server{})

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// GiveRecommendation returns recommendations within a given requirement.
func (s *Server) GetRecommendations(ctx context.Context, req *pb.Request) (*pb.Result, context.Context, error) {
	res := new(pb.Result)
	log.Trace().Msgf("GetRecommendations")
	require := req.Require
	if require == "dis" {
		p1 := &geoindex.GeoPoint{
			Pid:  "",
			Plat: req.Lat,
			Plon: req.Lon,
		}
		min := math.MaxFloat64
		for _, hotel := range s.hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp < min {
				min = tmp
			}
		}
		for _, hotel := range s.hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else if require == "rate" {
		max := 0.0
		for _, hotel := range s.hotels {
			if hotel.HRate > max {
				max = hotel.HRate
			}
		}
		for _, hotel := range s.hotels {
			if hotel.HRate == max {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else if require == "price" {
		min := math.MaxFloat64
		for _, hotel := range s.hotels {
			if hotel.HPrice < min {
				min = hotel.HPrice
			}
		}
		for _, hotel := range s.hotels {
			if hotel.HPrice == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else {
		log.Warn().Msgf("Wrong require parameter: %v", require)
	}

	return res, ctx, nil
}

// loadRecommendations loads hotel recommendations from mongodb.
func loadRecommendations(session *mgo.Session) map[string]Hotel {
	// session, err := mgo.Dial("mongodb-recommendation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	s := session.Copy()
	defer s.Close()

	c := s.DB("recommendation-db").C("recommendation")

	// unmarshal json profiles
	var hotels []Hotel
	err := c.Find(bson.M{}).All(&hotels)
	if err != nil {
		log.Error().Msgf("Failed get hotels data: ", err)
	}

	profiles := make(map[string]Hotel)
	for _, hotel := range hotels {
		profiles[hotel.HId] = hotel
	}

	return profiles
}

type Hotel struct {
	ID     bson.ObjectId `bson:"_id"`
	HId    string        `bson:"hotelId"`
	HLat   float64       `bson:"lat"`
	HLon   float64       `bson:"lon"`
	HRate  float64       `bson:"rate"`
	HPrice float64       `bson:"price"`
}
