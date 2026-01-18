package messagelogger

import (
	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	fb "github.com/appnetorg/hotel-reservation-arpc/proto/hotel/hotel_reservation"
	flatbuffers "github.com/google/flatbuffers/go"
)

// FlatBuffers converters for all hotel-reservation message types

// Geo service converters

func ProtoToFB_NearbyRequest(pbMsg *pb.NearbyRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	latstring := builder.CreateString(pbMsg.GetLatstring())

	fb.NearbyRequestStart(builder)
	fb.NearbyRequestAddLat(builder, pbMsg.GetLat())
	fb.NearbyRequestAddLon(builder, pbMsg.GetLon())
	fb.NearbyRequestAddLatstring(builder, latstring)
	obj := fb.NearbyRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_NearbyResult(pbMsg *pb.NearbyResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build hotel IDs vector
	hotelIds := pbMsg.GetHotelIds()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.NearbyResultStartHotelIdsVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.NearbyResultStart(builder)
	fb.NearbyResultAddHotelIds(builder, hotelIdsVector)
	obj := fb.NearbyResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

// Profile service converters

func ProtoToFB_GetProfilesRequest(pbMsg *pb.GetProfilesRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	locale := builder.CreateString(pbMsg.GetLocale())

	// Build hotel IDs vector
	hotelIds := pbMsg.GetHotelIds()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.GetProfilesRequestStartHotelIdsVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.GetProfilesRequestStart(builder)
	fb.GetProfilesRequestAddHotelIds(builder, hotelIdsVector)
	fb.GetProfilesRequestAddLocale(builder, locale)
	obj := fb.GetProfilesRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetProfilesResult(pbMsg *pb.GetProfilesResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(4096)

	// Build hotels list
	hotels := pbMsg.GetHotels()
	hotelOffsets := make([]flatbuffers.UOffsetT, len(hotels))

	for i, hotel := range hotels {
		hotelOffsets[i] = buildFBHotel(builder, hotel)
	}

	fb.GetProfilesResultStartHotelsVector(builder, len(hotelOffsets))
	for i := len(hotelOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelOffsets[i])
	}
	hotelsVector := builder.EndVector(len(hotelOffsets))

	fb.GetProfilesResultStart(builder)
	fb.GetProfilesResultAddHotels(builder, hotelsVector)
	obj := fb.GetProfilesResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Hotel(pbMsg *pb.Hotel) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	obj := buildFBHotel(builder, pbMsg)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func buildFBHotel(builder *flatbuffers.Builder, pbMsg *pb.Hotel) flatbuffers.UOffsetT {
	id := builder.CreateString(pbMsg.GetId())
	name := builder.CreateString(pbMsg.GetName())
	phoneNumber := builder.CreateString(pbMsg.GetPhoneNumber())
	description := builder.CreateString(pbMsg.GetDescription())

	// Build Address
	var addressOffset flatbuffers.UOffsetT
	if addr := pbMsg.GetAddress(); addr != nil {
		addressOffset = buildFBAddress(builder, addr)
	}

	// Build Images
	images := pbMsg.GetImages()
	imageOffsets := make([]flatbuffers.UOffsetT, len(images))
	for i, img := range images {
		imageOffsets[i] = buildFBImage(builder, img)
	}
	fb.HotelStartImagesVector(builder, len(imageOffsets))
	for i := len(imageOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(imageOffsets[i])
	}
	imagesVector := builder.EndVector(len(imageOffsets))

	fb.HotelStart(builder)
	fb.HotelAddId(builder, id)
	fb.HotelAddName(builder, name)
	fb.HotelAddPhoneNumber(builder, phoneNumber)
	fb.HotelAddDescription(builder, description)
	if addressOffset != 0 {
		fb.HotelAddAddress(builder, addressOffset)
	}
	fb.HotelAddImages(builder, imagesVector)
	return fb.HotelEnd(builder)
}

func ProtoToFB_Address(pbMsg *pb.Address) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	obj := buildFBAddress(builder, pbMsg)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func buildFBAddress(builder *flatbuffers.Builder, pbMsg *pb.Address) flatbuffers.UOffsetT {
	streetNumber := builder.CreateString(pbMsg.GetStreetNumber())
	streetName := builder.CreateString(pbMsg.GetStreetName())
	city := builder.CreateString(pbMsg.GetCity())
	state := builder.CreateString(pbMsg.GetState())
	country := builder.CreateString(pbMsg.GetCountry())
	postalCode := builder.CreateString(pbMsg.GetPostalCode())

	fb.AddressStart(builder)
	fb.AddressAddStreetNumber(builder, streetNumber)
	fb.AddressAddStreetName(builder, streetName)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddPostalCode(builder, postalCode)
	fb.AddressAddLat(builder, pbMsg.GetLat())
	fb.AddressAddLon(builder, pbMsg.GetLon())
	return fb.AddressEnd(builder)
}

func ProtoToFB_Image(pbMsg *pb.Image) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	obj := buildFBImage(builder, pbMsg)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func buildFBImage(builder *flatbuffers.Builder, pbMsg *pb.Image) flatbuffers.UOffsetT {
	url := builder.CreateString(pbMsg.GetUrl())

	fb.ImageStart(builder)
	fb.ImageAddUrl(builder, url)
	fb.ImageAddDefault(builder, pbMsg.GetDefault())
	return fb.ImageEnd(builder)
}

// Recommendation service converters

func ProtoToFB_GetRecommendationsRequest(pbMsg *pb.GetRecommendationsRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	require := builder.CreateString(pbMsg.GetRequire())

	fb.GetRecommendationsRequestStart(builder)
	fb.GetRecommendationsRequestAddRequire(builder, require)
	fb.GetRecommendationsRequestAddLat(builder, pbMsg.GetLat())
	fb.GetRecommendationsRequestAddLon(builder, pbMsg.GetLon())
	obj := fb.GetRecommendationsRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetRecommendationsResult(pbMsg *pb.GetRecommendationsResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build hotel IDs vector
	hotelIds := pbMsg.GetHotelIds()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.GetRecommendationsResultStartHotelIdsVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.GetRecommendationsResultStart(builder)
	fb.GetRecommendationsResultAddHotelIds(builder, hotelIdsVector)
	obj := fb.GetRecommendationsResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

// Rate service converters

func ProtoToFB_GetRatesRequest(pbMsg *pb.GetRatesRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	inDate := builder.CreateString(pbMsg.GetInDate())
	outDate := builder.CreateString(pbMsg.GetOutDate())

	// Build hotel IDs vector
	hotelIds := pbMsg.GetHotelIds()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.GetRatesRequestStartHotelIdsVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.GetRatesRequestStart(builder)
	fb.GetRatesRequestAddHotelIds(builder, hotelIdsVector)
	fb.GetRatesRequestAddInDate(builder, inDate)
	fb.GetRatesRequestAddOutDate(builder, outDate)
	obj := fb.GetRatesRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetRatesResult(pbMsg *pb.GetRatesResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(2048)

	// Build rate plans list
	ratePlans := pbMsg.GetRatePlans()
	ratePlanOffsets := make([]flatbuffers.UOffsetT, len(ratePlans))

	for i, rp := range ratePlans {
		ratePlanOffsets[i] = buildFBRatePlan(builder, rp)
	}

	fb.GetRatesResultStartRatePlansVector(builder, len(ratePlanOffsets))
	for i := len(ratePlanOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(ratePlanOffsets[i])
	}
	ratePlansVector := builder.EndVector(len(ratePlanOffsets))

	fb.GetRatesResultStart(builder)
	fb.GetRatesResultAddRatePlans(builder, ratePlansVector)
	obj := fb.GetRatesResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_RatePlan(pbMsg *pb.RatePlan) ([]byte, error) {
	builder := flatbuffers.NewBuilder(512)

	obj := buildFBRatePlan(builder, pbMsg)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func buildFBRatePlan(builder *flatbuffers.Builder, pbMsg *pb.RatePlan) flatbuffers.UOffsetT {
	hotelId := builder.CreateString(pbMsg.GetHotelId())
	code := builder.CreateString(pbMsg.GetCode())
	inDate := builder.CreateString(pbMsg.GetInDate())
	outDate := builder.CreateString(pbMsg.GetOutDate())

	// Build RoomType
	var roomTypeOffset flatbuffers.UOffsetT
	if rt := pbMsg.GetRoomType(); rt != nil {
		roomTypeOffset = buildFBRoomType(builder, rt)
	}

	fb.RatePlanStart(builder)
	fb.RatePlanAddHotelId(builder, hotelId)
	fb.RatePlanAddCode(builder, code)
	fb.RatePlanAddInDate(builder, inDate)
	fb.RatePlanAddOutDate(builder, outDate)
	if roomTypeOffset != 0 {
		fb.RatePlanAddRoomType(builder, roomTypeOffset)
	}
	return fb.RatePlanEnd(builder)
}

func ProtoToFB_RoomType(pbMsg *pb.RoomType) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	obj := buildFBRoomType(builder, pbMsg)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func buildFBRoomType(builder *flatbuffers.Builder, pbMsg *pb.RoomType) flatbuffers.UOffsetT {
	code := builder.CreateString(pbMsg.GetCode())
	currency := builder.CreateString(pbMsg.GetCurrency())
	roomDescription := builder.CreateString(pbMsg.GetRoomDescription())

	fb.RoomTypeStart(builder)
	fb.RoomTypeAddBookableRate(builder, pbMsg.GetBookableRate())
	fb.RoomTypeAddTotalRate(builder, pbMsg.GetTotalRate())
	fb.RoomTypeAddTotalRateInclusive(builder, pbMsg.GetTotalRateInclusive())
	fb.RoomTypeAddCode(builder, code)
	fb.RoomTypeAddCurrency(builder, currency)
	fb.RoomTypeAddRoomDescription(builder, roomDescription)
	return fb.RoomTypeEnd(builder)
}

// Reservation service converters

func ProtoToFB_ReservationRequest(pbMsg *pb.ReservationRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	customerName := builder.CreateString(pbMsg.GetCustomerName())
	inDate := builder.CreateString(pbMsg.GetInDate())
	outDate := builder.CreateString(pbMsg.GetOutDate())

	// Build hotel ID vector
	hotelIds := pbMsg.GetHotelId()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.ReservationRequestStartHotelIdVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.ReservationRequestStart(builder)
	fb.ReservationRequestAddCustomerName(builder, customerName)
	fb.ReservationRequestAddHotelId(builder, hotelIdsVector)
	fb.ReservationRequestAddInDate(builder, inDate)
	fb.ReservationRequestAddOutDate(builder, outDate)
	fb.ReservationRequestAddRoomNumber(builder, pbMsg.GetRoomNumber())
	obj := fb.ReservationRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ReservationResult(pbMsg *pb.ReservationResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build hotel ID vector
	hotelIds := pbMsg.GetHotelId()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.ReservationResultStartHotelIdVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.ReservationResultStart(builder)
	fb.ReservationResultAddHotelId(builder, hotelIdsVector)
	obj := fb.ReservationResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

// Search service converters

func ProtoToFB_SearchRequest(pbMsg *pb.SearchRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	inDate := builder.CreateString(pbMsg.GetInDate())
	outDate := builder.CreateString(pbMsg.GetOutDate())

	fb.SearchRequestStart(builder)
	fb.SearchRequestAddLat(builder, pbMsg.GetLat())
	fb.SearchRequestAddLon(builder, pbMsg.GetLon())
	fb.SearchRequestAddInDate(builder, inDate)
	fb.SearchRequestAddOutDate(builder, outDate)
	obj := fb.SearchRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_SearchResult(pbMsg *pb.SearchResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build hotel IDs vector
	hotelIds := pbMsg.GetHotelIds()
	hotelIdOffsets := make([]flatbuffers.UOffsetT, len(hotelIds))
	for i, id := range hotelIds {
		hotelIdOffsets[i] = builder.CreateString(id)
	}
	fb.SearchResultStartHotelIdsVector(builder, len(hotelIdOffsets))
	for i := len(hotelIdOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(hotelIdOffsets[i])
	}
	hotelIdsVector := builder.EndVector(len(hotelIdOffsets))

	fb.SearchResultStart(builder)
	fb.SearchResultAddHotelIds(builder, hotelIdsVector)
	obj := fb.SearchResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

// User service converters

func ProtoToFB_CheckUserRequest(pbMsg *pb.CheckUserRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	username := builder.CreateString(pbMsg.GetUsername())
	password := builder.CreateString(pbMsg.GetPassword())

	fb.CheckUserRequestStart(builder)
	fb.CheckUserRequestAddUsername(builder, username)
	fb.CheckUserRequestAddPassword(builder, password)
	obj := fb.CheckUserRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_CheckUserResult(pbMsg *pb.CheckUserResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(64)

	fb.CheckUserResultStart(builder)
	fb.CheckUserResultAddCorrect(builder, pbMsg.GetCorrect())
	obj := fb.CheckUserResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}
