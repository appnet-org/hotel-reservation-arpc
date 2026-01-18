package rate

import (
	"encoding/json"
	"fmt"
	"strconv"

	"context"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "os"
	"sort"
	"sync"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	"github.com/appnet-org/arpc/pkg/serializer"
	"github.com/rs/zerolog/log"

	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	"github.com/appnetorg/hotel-reservation-arpc/services/messagelogger"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"

	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

const _ = "srv-rate"

// Server implements the rate service
type Server struct {
	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	MemcClient   *memcache.Client
	uuid         string
}

// Run starts the server
func (s *Server) Run() error {
	opentracing.SetGlobalTracer(s.Tracer)

	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	serializer := &serializer.SymphonySerializer{}
	serverLogger, _ := messagelogger.NewServerMessageLogger("rate")
	server, err := rpc.NewServer(s.IpAddr+":"+strconv.Itoa(s.Port), serializer, []element.RPCElement{serverLogger})

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
		return err
	}

	pb.RegisterRateServer(server, s)

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// GetRates gets rates for hotels for specific date range.
func (s *Server) GetRates(ctx context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResult, context.Context, error) {
	// res := new(pb.Result)
	// session, err := mgo.Dial("mongodb-rate")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	ratePlans := make(RatePlans, 0)

	hotelIds := []string{}
	rateMap := make(map[string]struct{})
	for _, hotelID := range req.HotelIds {
		hotelIds = append(hotelIds, hotelID)
		rateMap[hotelID] = struct{}{}
	}
	// first check memcached(get-multi)
	memSpan, _ := opentracing.StartSpanFromContext(ctx, "memcached_get_multi_rate")
	memSpan.SetTag("span.kind", "client")
	resMap, err := s.MemcClient.GetMulti(hotelIds)
	memSpan.Finish()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	if err != nil && err != memcache.ErrCacheMiss {
		log.Panic().Msgf("Memmcached error while trying to get hotel [id: %v]= %s", hotelIds, err)
	} else {
		for _, item := range resMap {
			_ = strings.Split(string(item.Value), "\n") // unused for now

			// for _, rateStr := range rateStrs {
			// 	if len(rateStr) != 0 {
			// 		log.Info().Msgf("rateStr: %v", rateStr)
			// 		rateP := new(RatePlan)
			// 		json.Unmarshal([]byte(rateStr), rateP)
			// 		ratePlans = append(ratePlans, rateP)
			// 	}
			// }
		}
		wg.Add(len(rateMap))
		for hotelId := range rateMap {
			go func(id string) {
				// memcached miss, set up mongo connection
				session := s.MongoSession.Copy()
				defer session.Close()
				c := session.DB("rate-db").C("inventory")
				memcStr := ""
				tmpRatePlans := make(RatePlans, 0)
				mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_rate")
				mongoSpan.SetTag("span.kind", "client")
				err := c.Find(&bson.M{"hotelId": id}).All(&tmpRatePlans)
				mongoSpan.Finish()
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v], but got error: %s", id, err.Error())
				} else {
					for _, r := range tmpRatePlans {
						mutex.Lock()
						ratePlans = append(ratePlans, r)
						mutex.Unlock()
						rateJson, err := json.Marshal(r)
						if err != nil {
							log.Error().Msgf("Failed to marshal plan [Code: %v] with error: %s", r.Code, err)
						}
						memcStr = memcStr + string(rateJson) + "\n"
					}
				}
				go s.MemcClient.Set(&memcache.Item{Key: id, Value: []byte(memcStr)})

				defer wg.Done()
			}(hotelId)
		}
	}
	wg.Wait()

	sort.Sort(ratePlans)

	var resultRatePlans []*pb.RatePlan
	for _, ratePlan := range ratePlans {
		resultRatePlans = append(resultRatePlans, &pb.RatePlan{
			HotelId: ratePlan.HotelId,
			Code:    ratePlan.Code,
			InDate:  ratePlan.InDate,
			OutDate: ratePlan.OutDate,
			RoomType: &pb.RoomType{
				BookableRate:       ratePlan.RoomType.BookableRate,
				TotalRate:          ratePlan.RoomType.TotalRate,
				TotalRateInclusive: ratePlan.RoomType.TotalRateInclusive,
				Code:               ratePlan.RoomType.Code,
				Currency:           "USD", // Example, adjust accordingly
				RoomDescription:    ratePlan.RoomType.RoomDescription,
			},
		})
	}

	// Construct the result message
	result := &pb.GetRatesResult{
		RatePlans: resultRatePlans,
	}

	return result, ctx, nil
}

type RoomType struct {
	BookableRate       float64 `bson:"bookableRate"`
	Code               string  `bson:"code"`
	RoomDescription    string  `bson:"roomDescription"`
	TotalRate          float64 `bson:"totalRate"`
	TotalRateInclusive float64 `bson:"totalRateInclusive"`
}

type RatePlan struct {
	HotelId  string    `bson:"hotelId"`
	Code     string    `bson:"code"`
	InDate   string    `bson:"inDate"`
	OutDate  string    `bson:"outDate"`
	RoomType *RoomType `bson:"roomType"`
}

type RatePlans []*RatePlan

func (r RatePlans) Len() int {
	return len(r)
}

func (r RatePlans) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RatePlans) Less(i, j int) bool {
	return r[i].RoomType.TotalRate > r[j].RoomType.TotalRate
}

// func (r RatePlans) Less(i, j int) bool {
// 	// Check if the slice is nil or index out of bounds
// 	if r == nil || i >= len(r) || j >= len(r) {
// 		return false
// 	}

// 	// Check if r[i] or r[j] is nil
// 	if r[i] == nil || r[j] == nil {
// 		// You can define a behavior if one is nil. For example, treat nil as "less"
// 		return r[j] != nil
// 	}

// 	// Check if RoomType is nil in either r[i] or r[j]
// 	if r[i].RoomType == nil || r[j].RoomType == nil {
// 		// Again, you can define a behavior if RoomType is nil. Example:
// 		return r[j].RoomType != nil
// 	}

// 	// Check if TotalRate is nil, though it seems unlikely that TotalRate (a float or int) would be nil,
// 	// but just in case, you can handle it
// 	// If TotalRate is a pointer type, add a nil check; otherwise, proceed with the comparison.
// 	return r[i].RoomType.TotalRate > r[j].RoomType.TotalRate
// }
