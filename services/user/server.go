package user

import (
	"crypto/sha256"
	// "encoding/json"
	"fmt"

	"context"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	pb "github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "io/ioutil"

	"github.com/rs/zerolog/log"
	// "os"
)

const name = "srv-user"

// Server implements the user service
type Server struct {
	users map[string]string

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

	if s.users == nil {
		s.users = loadUsers(s.MongoSession)
	}

	s.uuid = uuid.New().String()

	serializer := &serializer.SymphonySerializer{}
	server, err := rpc.NewServer(s.IpAddr, serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
	}

	pb.RegisterUserServer(server, &Server{})

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// CheckUser returns whether the username and password are correct.
func (s *Server) CheckUser(ctx context.Context, req *pb.CheckUserRequest) (*pb.CheckUserResult, context.Context, error) {
	res := new(pb.CheckUserResult)

	log.Trace().Msg("CheckUser")

	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)

	// session, err := mgo.Dial("mongodb-user")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	// c := session.DB("user-db").C("user")

	// user := User{}
	// err = c.Find(bson.M{"username": req.Username}).One(&user)
	// if err != nil {
	// 	panic(err)
	// }
	res.Correct = false
	if true_pass, found := s.users[req.Username]; found {
		res.Correct = pass == true_pass
	}

	// res.Correct = user.Password == pass

	log.Trace().Msgf("CheckUser %d", res.Correct)

	return res, ctx, nil
}

// loadUsers loads hotel users from mongodb.
func loadUsers(session *mgo.Session) map[string]string {
	// session, err := mgo.Dial("mongodb-user")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	s := session.Copy()
	defer s.Close()
	c := s.DB("user-db").C("user")

	// unmarshal json profiles
	var users []User
	err := c.Find(bson.M{}).All(&users)
	if err != nil {
		log.Error().Msgf("Failed get users data: ", err)
	}

	res := make(map[string]string)
	for _, user := range users {
		res[user.Username] = user.Password
	}

	log.Trace().Msg("Done load users")

	return res
}

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
