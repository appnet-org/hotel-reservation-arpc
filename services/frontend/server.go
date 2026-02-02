package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/appnet-org/arpc/pkg/metadata"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	hotel "github.com/appnetorg/hotel-reservation-arpc/proto"
	"github.com/appnetorg/hotel-reservation-arpc/tls"
	"github.com/rs/zerolog/log"

	"github.com/appnetorg/hotel-reservation-arpc/tracing"
	"github.com/opentracing/opentracing-go"
)

// Server implements frontend service
type Server struct {
	searchClient         hotel.SearchClient
	profileClient        hotel.ProfileClient
	recommendationClient hotel.RecommendationClient
	userClient           hotel.UserClient
	reservationClient    hotel.ReservationClient
	KnativeDns           string
	IpAddr               string
	Port                 int
	Tracer               opentracing.Tracer
}

// Run the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("Server port must be set")
	}

	if err := s.initSearchClient("search.default.svc.cluster.local:11002"); err != nil {
		return err
	}

	if err := s.initProfileClient("profile.default.svc.cluster.local:11001"); err != nil {
		return err
	}

	if err := s.initRecommendationClient("recommendation.default.svc.cluster.local:11005"); err != nil {
		return err
	}

	if err := s.initUserClient("user.default.svc.cluster.local:11006"); err != nil {
		return err
	}

	if err := s.initReservation("reservation.default.svc.cluster.local:11007"); err != nil {
		return err
	}

	mux := tracing.NewServeMux(s.Tracer)
	mux.Handle("/", http.FileServer(http.Dir("services/frontend/static")))
	mux.Handle("/hotels", http.HandlerFunc(s.searchHandler))
	mux.Handle("/recommendations", http.HandlerFunc(s.recommendHandler))
	mux.Handle("/user", http.HandlerFunc(s.userHandler))
	mux.Handle("/reservation", http.HandlerFunc(s.reservationHandler))

	tlsconfig := tls.GetHttpsOpt()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}
	if tlsconfig != nil {
		log.Info().Msg("Serving https")
		srv.TLSConfig = tlsconfig
		return srv.ListenAndServeTLS("x509/server_cert.pem", "x509/server_key.pem")
	} else {
		log.Info().Msg("Serving http")
		return srv.ListenAndServe()
	}
}

func (s *Server) initSearchClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create search aRPC client: %v", err)
	}

	s.searchClient = hotel.NewSearchClient(client)
	return nil
}

func (s *Server) initProfileClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create profile aRPC client: %v", err)
	}

	s.profileClient = hotel.NewProfileClient(client)
	return nil
}

func (s *Server) initRecommendationClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create recommendation aRPC client: %v", err)
	}
	s.recommendationClient = hotel.NewRecommendationClient(client)
	return nil
}

func (s *Server) initUserClient(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create user aRPC client: %v", err)
	}
	s.userClient = hotel.NewUserClient(client)
	return nil
}

func (s *Server) initReservation(name string) error {
	serializer := &serializer.SymphonySerializer{}

	client, err := rpc.NewClientWithLocalAddr(serializer, name, "0.0.0.0:0", nil)
	if err != nil {
		return fmt.Errorf("failed to create reservation aRPC client: %v", err)
	}
	s.reservationClient = hotel.NewReservationClient(client)
	return nil
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// ctx := r.Context()

	md := metadata.New(map[string]string{})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// lan/lon from query params
	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}

	Lat, _ := strconv.ParseFloat(sLat, 32)
	lat := float32(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 32)
	lon := float32(Lon)

	// search for best hotels
	searchResp, err := s.searchClient.Nearby(ctx, &hotel.SearchRequest{
		Lat:     lat,
		Lon:     lon,
		InDate:  inDate,
		OutDate: outDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	reservationResp, err := s.reservationClient.CheckAvailability(ctx, &hotel.ReservationRequest{
		CustomerName: "",
		HotelId:      searchResp.HotelIds,
		InDate:       inDate,
		OutDate:      outDate,
		RoomNumber:   1,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &hotel.GetProfilesRequest{
		HotelIds: reservationResp.HotelId,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}

func (s *Server) recommendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}
	Lat, _ := strconv.ParseFloat(sLat, 64)
	lat := float64(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 64)
	lon := float64(Lon)

	require := r.URL.Query().Get("require")
	if require != "dis" && require != "rate" && require != "price" {
		http.Error(w, "Please specify require params", http.StatusBadRequest)
		return
	}

	// recommend hotels
	recResp, err := s.recommendationClient.GetRecommendations(ctx, &hotel.GetRecommendationsRequest{
		Require: require,
		Lat:     float64(lat),
		Lon:     float64(lon),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &hotel.GetProfilesRequest{
		HotelIds: recResp.HotelIds,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}

func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// ctx := r.Context()

	md := metadata.New(map[string]string{})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &hotel.CheckUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Login successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) reservationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// ctx := r.Context()

	md := metadata.New(map[string]string{})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		http.Error(w, "Please check inDate/outDate format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	customerName := r.URL.Query().Get("customerName")
	if customerName == "" {
		http.Error(w, "Please specify customerName params", http.StatusBadRequest)
		return
	}

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	numberOfRoom := 0
	num := r.URL.Query().Get("number")
	if num != "" {
		numberOfRoom, _ = strconv.Atoi(num)
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &hotel.CheckUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Reserve successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	// Make reservation
	resResp, err := s.reservationClient.MakeReservation(ctx, &hotel.ReservationRequest{
		CustomerName: customerName,
		HotelId:      []string{hotelId},
		InDate:       inDate,
		OutDate:      outDate,
		RoomNumber:   int32(numberOfRoom),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resResp.HotelId) == 0 {
		str = "Failed. Already reserved. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

// return a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSONResponse(hs []*hotel.Hotel) map[string]interface{} {
	fs := []interface{}{}

	for _, h := range hs {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float32{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}

func checkDataFormat(date string) bool {
	if len(date) != 10 {
		return false
	}
	for i := 0; i < 10; i++ {
		if i == 4 || i == 7 {
			if date[i] != '-' {
				return false
			}
		} else {
			if date[i] < '0' || date[i] > '9' {
				return false
			}
		}
	}
	return true
}
