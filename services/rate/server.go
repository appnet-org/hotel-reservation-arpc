package rate

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "io/ioutil"
	"net"
	// "os"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	interceptor "github.com/appnet-org/go-lib/interceptor"
	pb "github.com/appnetorg/HotelReservation/services/rate/proto"
	"github.com/appnetorg/HotelReservation/tls"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

const name = "srv-rate"

// Server implements the rate service
type Server struct {
	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	MemcClient   *memcache.Client
	uuid         string
	pb.UnimplementedRateServer
}

// Run starts the server
func (s *Server) Run() error {
	opentracing.SetGlobalTracer(s.Tracer)

	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: 120 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			// otgrpc.OpenTracingServerInterceptor(s.Tracer),
			interceptor.ServerInterceptor("/appnet/interceptors/rate"),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterRateServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}

	// register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := ioutil.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// GetRates gets rates for hotels for specific date range.
func (s *Server) GetRates(ctx context.Context, req *pb.Request) (*pb.Result, error) {
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
	log.Trace().Msgf("resMap = %v", resMap)
	memSpan.Finish()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	if err != nil && err != memcache.ErrCacheMiss {
		log.Panic().Msgf("Memmcached error while trying to get hotel [id: %v]= %s", hotelIds, err)
	} else {
		for hotelId, item := range resMap {
			rateStrs := strings.Split(string(item.Value), "\n")
			log.Trace().Msgf("memc hit, hotelId = %s,rate strings: %v", hotelId, rateStrs)

			// for _, rateStr := range rateStrs {
			// 	if len(rateStr) != 0 {
			// 		log.Info().Msgf("rateStr: %v", rateStr)
			// 		rateP := new(RatePlan)
			// 		json.Unmarshal([]byte(rateStr), rateP)
			// 		ratePlans = append(ratePlans, rateP)
			// 	}
			// }
			// delete(rateMap, hotelId)
		}
		wg.Add(len(rateMap))
		for hotelId := range rateMap {
			go func(id string) {
				log.Trace().Msgf("memc miss, hotelId = %s", id)
				log.Trace().Msg("memcached miss, set up mongo connection")

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
				log.Trace().Msgf("tmpRatePlans = %v", tmpRatePlans)
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v], but got error", id, err.Error())
				} else {
					for _, r := range tmpRatePlans {
						log.Trace().Msgf("RatePlan HotelId = %s, Code = %s", r.HotelId, r.Code)
						log.Trace().Msgf("RatePlan RoomType = %v", r.RoomType)
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

	// Printing the ratePlans
	for _, ratePlan := range ratePlans {
		log.Trace().Msgf("RatePlan HotelId = %s, Code = %s", ratePlan.HotelId, ratePlan.Code)
	}

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
	result := &pb.Result{
		RatePlans: resultRatePlans,
	}

	return result, nil
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
