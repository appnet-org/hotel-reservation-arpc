package profile

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "io/ioutil"

	// "os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	pb "github.com/appnetorg/hotel-reservation-arpc/services/profile/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"

	"github.com/bradfitz/gomemcache/memcache"
	// "strings"
)

const name = "srv-profile"

// Server implements the profile service
type Server struct {
	Tracer       opentracing.Tracer
	uuid         string
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	MemcClient   *memcache.Client
}

// Run starts the server
func (s *Server) Run() error {
	opentracing.SetGlobalTracer(s.Tracer)

	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	log.Trace().Msgf("in run s.IpAddr = %s, port = %d", s.IpAddr, s.Port)

	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer(s.IpAddr, serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterProfileServer(server, &Server{})

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// GetProfiles returns hotel profiles for requested IDs
func (s *Server) GetProfiles(ctx context.Context, req *pb.Request) (*pb.Result, context.Context, error) {
	// session, err := mgo.Dial("mongodb-profile")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	log.Trace().Msgf("In GetProfiles")

	hotels := make([]*pb.Result, 0)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// one hotel should only have one profile
	hotelIds := make([]string, 0)
	profileMap := make(map[string]struct{})
	for _, hotelId := range req.HotelIds {
		hotelIds = append(hotelIds, hotelId)
		profileMap[hotelId] = struct{}{}
	}

	log.Trace().Msgf("length of hotelIds: %v", len(hotelIds))

	memSpan, _ := opentracing.StartSpanFromContext(ctx, "memcached_get_profile")
	memSpan.SetTag("span.kind", "client")
	resMap, err := s.MemcClient.GetMulti(hotelIds)
	memSpan.Finish()
	if err != nil && err != memcache.ErrCacheMiss {
		log.Panic().Msgf("Tried to get hotelIds [%v], but got memmcached error = %s", hotelIds, err)
	} else {
		for hotelId, item := range resMap {
			profileStr := string(item.Value)
			log.Trace().Msgf("memc hit with %v", profileStr)

			hotelProf := new(pb.Result)
			json.Unmarshal(item.Value, hotelProf)
			hotels = append(hotels, hotelProf)
			delete(profileMap, hotelId)
		}

		wg.Add(len(profileMap))
		for hotelId := range profileMap {
			go func(hotelId string) {
				session := s.MongoSession.Copy()
				defer session.Close()
				c := session.DB("profile-db").C("hotels")

				hotelProf := new(pb.Result)
				mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_profile")
				mongoSpan.SetTag("span.kind", "client")
				err := c.Find(bson.M{"id": hotelId}).One(&hotelProf)
				mongoSpan.Finish()

				if err != nil {
					log.Error().Msgf("Failed get hotels data: %v", err)
				}

				mutex.Lock()
				hotels = append(hotels, hotelProf)
				mutex.Unlock()

				profJson, err := json.Marshal(hotelProf)
				if err != nil {
					log.Error().Msgf("Failed to marshal hotel [id: %v] with err: %v", hotelProf.Id, err)
				}
				memcStr := string(profJson)

				// write to memcached
				go s.MemcClient.Set(&memcache.Item{Key: hotelId, Value: []byte(memcStr)})
				defer wg.Done()
			}(hotelId)
		}
	}
	wg.Wait()

	log.Trace().Msgf("In GetProfiles after getting resp")
	return hotels[0], ctx, nil
}
