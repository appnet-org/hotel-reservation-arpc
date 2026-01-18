package messagelogger

import (
	"encoding/json"
	"fmt"
	"reflect"

	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	"google.golang.org/protobuf/proto"
)

// SerializationSizes holds the sizes of different serialization formats
type SerializationSizes struct {
	Protobuf       int               `json:"protobuf"`
	FlatBuffers    int               `json:"flatbuffers"`
	CapnProto      int               `json:"capnproto"`
	Symphony       int               `json:"symphony"`
	SymphonyHybrid int               `json:"symphony_hybrid"`
	Errors         map[string]string `json:"errors,omitempty"`
}

// ComputeSizes computes serialization sizes for all formats
func ComputeSizes(msg interface{}) SerializationSizes {
	sizes := SerializationSizes{
		Protobuf:       -1,
		FlatBuffers:    -1,
		CapnProto:      -1,
		Symphony:       -1,
		SymphonyHybrid: -1,
		Errors:         make(map[string]string),
	}

	// Handle nil messages early
	if msg == nil {
		sizes.Errors["protobuf"] = "message is nil"
		sizes.Errors["flatbuffers"] = "message is nil"
		sizes.Errors["capnproto"] = "message is nil"
		sizes.Errors["symphony"] = "message is nil"
		sizes.Errors["symphony_hybrid"] = "message is nil"
		return sizes
	}

	// Try to compute protobuf size
	if protoMsg, ok := msg.(proto.Message); ok {
		data, err := proto.Marshal(protoMsg)
		if err != nil {
			sizes.Errors["protobuf"] = err.Error()
		} else {
			sizes.Protobuf = len(data)
		}
	} else {
		sizes.Errors["protobuf"] = "not a proto message"
	}

	// Compute FlatBuffers size
	fbData, fbErr := convertToFlatBuffers(msg)
	if fbErr != nil {
		sizes.Errors["flatbuffers"] = fbErr.Error()
	} else {
		sizes.FlatBuffers = len(fbData)
	}

	// Compute Cap'n Proto size
	capnpData, capnpErr := convertToCapnProto(msg)
	if capnpErr != nil {
		sizes.Errors["capnproto"] = capnpErr.Error()
	} else {
		sizes.CapnProto = len(capnpData)
	}

	// Compute Symphony size
	symphonyData, symphonyErr := convertToSymphony(msg)
	if symphonyErr != nil {
		sizes.Errors["symphony"] = symphonyErr.Error()
	} else {
		// Subtract 9 bytes because Symphony encodes extra information
		symphonySize := len(symphonyData) - 9
		if symphonySize < 0 {
			symphonySize = 0
		}
		sizes.Symphony = symphonySize
	}

	// Compute Symphony Hybrid size
	symphonyHybridData, symphonyHybridErr := convertToSymphonyHybrid(msg)
	if symphonyHybridErr != nil {
		sizes.Errors["symphony_hybrid"] = symphonyHybridErr.Error()
	} else {
		symphonyHybridSize := len(symphonyHybridData) - 9
		if symphonyHybridSize < 0 {
			symphonyHybridSize = 0
		}
		sizes.SymphonyHybrid = symphonyHybridSize
	}

	return sizes
}

// convertToFlatBuffers converts a protobuf message to FlatBuffers
func convertToFlatBuffers(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	typeName := reflect.TypeOf(msg).String()

	switch v := msg.(type) {
	// Geo service
	case *pb.NearbyRequest:
		return ProtoToFB_NearbyRequest(v)
	case *pb.NearbyResult:
		return ProtoToFB_NearbyResult(v)

	// Profile service
	case *pb.GetProfilesRequest:
		return ProtoToFB_GetProfilesRequest(v)
	case *pb.GetProfilesResult:
		return ProtoToFB_GetProfilesResult(v)
	case *pb.Hotel:
		return ProtoToFB_Hotel(v)
	case *pb.Address:
		return ProtoToFB_Address(v)
	case *pb.Image:
		return ProtoToFB_Image(v)

	// Recommendation service
	case *pb.GetRecommendationsRequest:
		return ProtoToFB_GetRecommendationsRequest(v)
	case *pb.GetRecommendationsResult:
		return ProtoToFB_GetRecommendationsResult(v)

	// Rate service
	case *pb.GetRatesRequest:
		return ProtoToFB_GetRatesRequest(v)
	case *pb.GetRatesResult:
		return ProtoToFB_GetRatesResult(v)
	case *pb.RatePlan:
		return ProtoToFB_RatePlan(v)
	case *pb.RoomType:
		return ProtoToFB_RoomType(v)

	// Reservation service
	case *pb.ReservationRequest:
		return ProtoToFB_ReservationRequest(v)
	case *pb.ReservationResult:
		return ProtoToFB_ReservationResult(v)

	// Search service
	case *pb.SearchRequest:
		return ProtoToFB_SearchRequest(v)
	case *pb.SearchResult:
		return ProtoToFB_SearchResult(v)

	// User service
	case *pb.CheckUserRequest:
		return ProtoToFB_CheckUserRequest(v)
	case *pb.CheckUserResult:
		return ProtoToFB_CheckUserResult(v)

	default:
		return nil, fmt.Errorf("unsupported message type for FlatBuffers conversion: %s", typeName)
	}
}

// convertToCapnProto converts a protobuf message to Cap'n Proto
func convertToCapnProto(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	typeName := reflect.TypeOf(msg).String()

	switch v := msg.(type) {
	// Geo service
	case *pb.NearbyRequest:
		return ProtoToCapnp_NearbyRequest(v)
	case *pb.NearbyResult:
		return ProtoToCapnp_NearbyResult(v)

	// Profile service
	case *pb.GetProfilesRequest:
		return ProtoToCapnp_GetProfilesRequest(v)
	case *pb.GetProfilesResult:
		return ProtoToCapnp_GetProfilesResult(v)
	case *pb.Hotel:
		return ProtoToCapnp_Hotel(v)
	case *pb.Address:
		return ProtoToCapnp_Address(v)
	case *pb.Image:
		return ProtoToCapnp_Image(v)

	// Recommendation service
	case *pb.GetRecommendationsRequest:
		return ProtoToCapnp_GetRecommendationsRequest(v)
	case *pb.GetRecommendationsResult:
		return ProtoToCapnp_GetRecommendationsResult(v)

	// Rate service
	case *pb.GetRatesRequest:
		return ProtoToCapnp_GetRatesRequest(v)
	case *pb.GetRatesResult:
		return ProtoToCapnp_GetRatesResult(v)
	case *pb.RatePlan:
		return ProtoToCapnp_RatePlan(v)
	case *pb.RoomType:
		return ProtoToCapnp_RoomType(v)

	// Reservation service
	case *pb.ReservationRequest:
		return ProtoToCapnp_ReservationRequest(v)
	case *pb.ReservationResult:
		return ProtoToCapnp_ReservationResult(v)

	// Search service
	case *pb.SearchRequest:
		return ProtoToCapnp_SearchRequest(v)
	case *pb.SearchResult:
		return ProtoToCapnp_SearchResult(v)

	// User service
	case *pb.CheckUserRequest:
		return ProtoToCapnp_CheckUserRequest(v)
	case *pb.CheckUserResult:
		return ProtoToCapnp_CheckUserResult(v)

	default:
		return nil, fmt.Errorf("unsupported message type for Cap'n Proto conversion: %s", typeName)
	}
}

// convertToSymphony converts a protobuf message to Symphony
func convertToSymphony(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	// Define interface for Symphony marshallable types
	type symphonyMarshaller interface {
		MarshalSymphony() ([]byte, error)
	}

	// Check if the message implements MarshalSymphony
	if sm, ok := msg.(symphonyMarshaller); ok {
		return sm.MarshalSymphony()
	}

	typeName := reflect.TypeOf(msg).String()
	return nil, fmt.Errorf("unsupported message type for Symphony conversion: %s", typeName)
}

// convertToSymphonyHybrid converts a protobuf message to Symphony Hybrid
func convertToSymphonyHybrid(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	type symphonyHybridMarshaller interface {
		MarshalSymphonyHybrid() ([]byte, error)
	}

	if shm, ok := msg.(symphonyHybridMarshaller); ok {
		return shm.MarshalSymphonyHybrid()
	}

	typeName := reflect.TypeOf(msg).String()
	return nil, fmt.Errorf("unsupported message type for Symphony Hybrid conversion: %s", typeName)
}

// GetMessageTypeName returns the type name of the message
func GetMessageTypeName(msg interface{}) string {
	if msg == nil {
		return "unknown"
	}

	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

// LogEntry represents a structured log entry with serialization sizes
type LogEntry struct {
	Timestamp   string             `json:"timestamp"`
	Direction   string             `json:"direction"` // "request" or "response"
	Method      string             `json:"method,omitempty"`
	MessageType string             `json:"message_type"`
	Sizes       SerializationSizes `json:"sizes"`
	Payload     interface{}        `json:"payload"`
}

// MarshalLogEntry marshals a log entry to JSON
func MarshalLogEntry(entry LogEntry) ([]byte, error) {
	return json.Marshal(entry)
}
