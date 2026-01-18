package messagelogger

import (
	"fmt"

	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	pbcapnp "github.com/appnetorg/hotel-reservation-arpc/proto/hotel/capnp"
	capnp "capnproto.org/go/capnp/v3"
)

// Cap'n Proto converters for all hotel-reservation message types

// Geo service converters

func ProtoToCapnp_NearbyRequest(pbMsg *pb.NearbyRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootNearbyRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create NearbyRequest: %w", err)
	}

	req.SetLat(pbMsg.GetLat())
	req.SetLon(pbMsg.GetLon())
	if err := req.SetLatstring(pbMsg.GetLatstring()); err != nil {
		return nil, fmt.Errorf("failed to set latstring: %w", err)
	}

	return msg.Marshal()
}

func ProtoToCapnp_NearbyResult(pbMsg *pb.NearbyResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootNearbyResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create NearbyResult: %w", err)
	}

	hotelIds := pbMsg.GetHotelIds()
	list, err := res.NewHotelIds(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

// Profile service converters

func ProtoToCapnp_GetProfilesRequest(pbMsg *pb.GetProfilesRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetProfilesRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetProfilesRequest: %w", err)
	}

	hotelIds := pbMsg.GetHotelIds()
	list, err := req.NewHotelIds(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	if err := req.SetLocale(pbMsg.GetLocale()); err != nil {
		return nil, fmt.Errorf("failed to set locale: %w", err)
	}

	return msg.Marshal()
}

func ProtoToCapnp_GetProfilesResult(pbMsg *pb.GetProfilesResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootGetProfilesResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetProfilesResult: %w", err)
	}

	hotels := pbMsg.GetHotels()
	hotelList, err := res.NewHotels(int32(len(hotels)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotels list: %w", err)
	}
	for i, h := range hotels {
		hotel := hotelList.At(i)
		if err := setCapnpHotel(hotel, h); err != nil {
			return nil, fmt.Errorf("failed to set hotel %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

func ProtoToCapnp_Hotel(pbMsg *pb.Hotel) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	hotel, err := pbcapnp.NewRootHotel(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Hotel: %w", err)
	}

	if err := setCapnpHotel(hotel, pbMsg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func setCapnpHotel(hotel pbcapnp.Hotel, pbMsg *pb.Hotel) error {
	if err := hotel.SetId(pbMsg.GetId()); err != nil {
		return fmt.Errorf("failed to set id: %w", err)
	}
	if err := hotel.SetName(pbMsg.GetName()); err != nil {
		return fmt.Errorf("failed to set name: %w", err)
	}
	if err := hotel.SetPhoneNumber(pbMsg.GetPhoneNumber()); err != nil {
		return fmt.Errorf("failed to set phone_number: %w", err)
	}
	if err := hotel.SetDescription(pbMsg.GetDescription()); err != nil {
		return fmt.Errorf("failed to set description: %w", err)
	}

	// Set Address
	if addr := pbMsg.GetAddress(); addr != nil {
		capnpAddr, err := hotel.NewAddress()
		if err != nil {
			return fmt.Errorf("failed to create address: %w", err)
		}
		if err := setCapnpAddress(capnpAddr, addr); err != nil {
			return err
		}
	}

	// Set Images
	images := pbMsg.GetImages()
	imageList, err := hotel.NewImages(int32(len(images)))
	if err != nil {
		return fmt.Errorf("failed to create images list: %w", err)
	}
	for i, img := range images {
		capnpImg := imageList.At(i)
		if err := setCapnpImage(capnpImg, img); err != nil {
			return fmt.Errorf("failed to set image %d: %w", i, err)
		}
	}

	return nil
}

func ProtoToCapnp_Address(pbMsg *pb.Address) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	addr, err := pbcapnp.NewRootAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}

	if err := setCapnpAddress(addr, pbMsg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func setCapnpAddress(addr pbcapnp.Address, pbMsg *pb.Address) error {
	if err := addr.SetStreetNumber(pbMsg.GetStreetNumber()); err != nil {
		return fmt.Errorf("failed to set street_number: %w", err)
	}
	if err := addr.SetStreetName(pbMsg.GetStreetName()); err != nil {
		return fmt.Errorf("failed to set street_name: %w", err)
	}
	if err := addr.SetCity(pbMsg.GetCity()); err != nil {
		return fmt.Errorf("failed to set city: %w", err)
	}
	if err := addr.SetState(pbMsg.GetState()); err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}
	if err := addr.SetCountry(pbMsg.GetCountry()); err != nil {
		return fmt.Errorf("failed to set country: %w", err)
	}
	if err := addr.SetPostalCode(pbMsg.GetPostalCode()); err != nil {
		return fmt.Errorf("failed to set postal_code: %w", err)
	}
	addr.SetLat(pbMsg.GetLat())
	addr.SetLon(pbMsg.GetLon())
	return nil
}

func ProtoToCapnp_Image(pbMsg *pb.Image) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	img, err := pbcapnp.NewRootImage(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Image: %w", err)
	}

	if err := setCapnpImage(img, pbMsg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func setCapnpImage(img pbcapnp.Image, pbMsg *pb.Image) error {
	if err := img.SetUrl(pbMsg.GetUrl()); err != nil {
		return fmt.Errorf("failed to set url: %w", err)
	}
	img.SetDefault(pbMsg.GetDefault())
	return nil
}

// Recommendation service converters

func ProtoToCapnp_GetRecommendationsRequest(pbMsg *pb.GetRecommendationsRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetRecommendationsRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetRecommendationsRequest: %w", err)
	}

	if err := req.SetRequire(pbMsg.GetRequire()); err != nil {
		return nil, fmt.Errorf("failed to set require: %w", err)
	}
	req.SetLat(pbMsg.GetLat())
	req.SetLon(pbMsg.GetLon())

	return msg.Marshal()
}

func ProtoToCapnp_GetRecommendationsResult(pbMsg *pb.GetRecommendationsResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootGetRecommendationsResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetRecommendationsResult: %w", err)
	}

	hotelIds := pbMsg.GetHotelIds()
	list, err := res.NewHotelIds(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

// Rate service converters

func ProtoToCapnp_GetRatesRequest(pbMsg *pb.GetRatesRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetRatesRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetRatesRequest: %w", err)
	}

	hotelIds := pbMsg.GetHotelIds()
	list, err := req.NewHotelIds(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	if err := req.SetInDate(pbMsg.GetInDate()); err != nil {
		return nil, fmt.Errorf("failed to set in_date: %w", err)
	}
	if err := req.SetOutDate(pbMsg.GetOutDate()); err != nil {
		return nil, fmt.Errorf("failed to set out_date: %w", err)
	}

	return msg.Marshal()
}

func ProtoToCapnp_GetRatesResult(pbMsg *pb.GetRatesResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootGetRatesResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetRatesResult: %w", err)
	}

	ratePlans := pbMsg.GetRatePlans()
	rpList, err := res.NewRatePlans(int32(len(ratePlans)))
	if err != nil {
		return nil, fmt.Errorf("failed to create rate plans list: %w", err)
	}
	for i, rp := range ratePlans {
		capnpRP := rpList.At(i)
		if err := setCapnpRatePlan(capnpRP, rp); err != nil {
			return nil, fmt.Errorf("failed to set rate plan %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

func ProtoToCapnp_RatePlan(pbMsg *pb.RatePlan) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	rp, err := pbcapnp.NewRootRatePlan(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create RatePlan: %w", err)
	}

	if err := setCapnpRatePlan(rp, pbMsg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func setCapnpRatePlan(rp pbcapnp.RatePlan, pbMsg *pb.RatePlan) error {
	if err := rp.SetHotelId(pbMsg.GetHotelId()); err != nil {
		return fmt.Errorf("failed to set hotel_id: %w", err)
	}
	if err := rp.SetCode(pbMsg.GetCode()); err != nil {
		return fmt.Errorf("failed to set code: %w", err)
	}
	if err := rp.SetInDate(pbMsg.GetInDate()); err != nil {
		return fmt.Errorf("failed to set in_date: %w", err)
	}
	if err := rp.SetOutDate(pbMsg.GetOutDate()); err != nil {
		return fmt.Errorf("failed to set out_date: %w", err)
	}

	// Set RoomType
	if rt := pbMsg.GetRoomType(); rt != nil {
		capnpRT, err := rp.NewRoomType()
		if err != nil {
			return fmt.Errorf("failed to create room_type: %w", err)
		}
		if err := setCapnpRoomType(capnpRT, rt); err != nil {
			return err
		}
	}

	return nil
}

func ProtoToCapnp_RoomType(pbMsg *pb.RoomType) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	rt, err := pbcapnp.NewRootRoomType(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create RoomType: %w", err)
	}

	if err := setCapnpRoomType(rt, pbMsg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func setCapnpRoomType(rt pbcapnp.RoomType, pbMsg *pb.RoomType) error {
	rt.SetBookableRate(pbMsg.GetBookableRate())
	rt.SetTotalRate(pbMsg.GetTotalRate())
	rt.SetTotalRateInclusive(pbMsg.GetTotalRateInclusive())
	if err := rt.SetCode(pbMsg.GetCode()); err != nil {
		return fmt.Errorf("failed to set code: %w", err)
	}
	if err := rt.SetCurrency(pbMsg.GetCurrency()); err != nil {
		return fmt.Errorf("failed to set currency: %w", err)
	}
	if err := rt.SetRoomDescription(pbMsg.GetRoomDescription()); err != nil {
		return fmt.Errorf("failed to set room_description: %w", err)
	}
	return nil
}

// Reservation service converters

func ProtoToCapnp_ReservationRequest(pbMsg *pb.ReservationRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootReservationRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ReservationRequest: %w", err)
	}

	if err := req.SetCustomerName(pbMsg.GetCustomerName()); err != nil {
		return nil, fmt.Errorf("failed to set customer_name: %w", err)
	}

	hotelIds := pbMsg.GetHotelId()
	list, err := req.NewHotelId(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	if err := req.SetInDate(pbMsg.GetInDate()); err != nil {
		return nil, fmt.Errorf("failed to set in_date: %w", err)
	}
	if err := req.SetOutDate(pbMsg.GetOutDate()); err != nil {
		return nil, fmt.Errorf("failed to set out_date: %w", err)
	}
	req.SetRoomNumber(pbMsg.GetRoomNumber())

	return msg.Marshal()
}

func ProtoToCapnp_ReservationResult(pbMsg *pb.ReservationResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootReservationResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ReservationResult: %w", err)
	}

	hotelIds := pbMsg.GetHotelId()
	list, err := res.NewHotelId(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

// Search service converters

func ProtoToCapnp_SearchRequest(pbMsg *pb.SearchRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootSearchRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SearchRequest: %w", err)
	}

	req.SetLat(pbMsg.GetLat())
	req.SetLon(pbMsg.GetLon())
	if err := req.SetInDate(pbMsg.GetInDate()); err != nil {
		return nil, fmt.Errorf("failed to set in_date: %w", err)
	}
	if err := req.SetOutDate(pbMsg.GetOutDate()); err != nil {
		return nil, fmt.Errorf("failed to set out_date: %w", err)
	}

	return msg.Marshal()
}

func ProtoToCapnp_SearchResult(pbMsg *pb.SearchResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootSearchResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SearchResult: %w", err)
	}

	hotelIds := pbMsg.GetHotelIds()
	list, err := res.NewHotelIds(int32(len(hotelIds)))
	if err != nil {
		return nil, fmt.Errorf("failed to create hotel IDs list: %w", err)
	}
	for i, id := range hotelIds {
		if err := list.Set(i, id); err != nil {
			return nil, fmt.Errorf("failed to set hotel ID %d: %w", i, err)
		}
	}

	return msg.Marshal()
}

// User service converters

func ProtoToCapnp_CheckUserRequest(pbMsg *pb.CheckUserRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootCheckUserRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CheckUserRequest: %w", err)
	}

	if err := req.SetUsername(pbMsg.GetUsername()); err != nil {
		return nil, fmt.Errorf("failed to set username: %w", err)
	}
	if err := req.SetPassword(pbMsg.GetPassword()); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	return msg.Marshal()
}

func ProtoToCapnp_CheckUserResult(pbMsg *pb.CheckUserResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res, err := pbcapnp.NewRootCheckUserResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CheckUserResult: %w", err)
	}

	res.SetCorrect(pbMsg.GetCorrect())

	return msg.Marshal()
}
