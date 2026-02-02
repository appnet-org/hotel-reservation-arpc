package user

import (
	"crypto/sha256"
	"strconv"

	// "encoding/json"
	"fmt"

	"context"

	"github.com/appnet-org/arpc/pkg/rpc"
	"github.com/appnet-org/arpc/pkg/serializer"
	pb "github.com/appnetorg/hotel-reservation-arpc/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/rs/zerolog/log"
)

const _ = "srv-user"

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
	server, err := rpc.NewServer(s.IpAddr+":"+strconv.Itoa(s.Port), serializer, nil)

	if err != nil {
		log.Error().Msgf("Failed to start aRPC server: %v", err)
		return err
	}

	pb.RegisterUserServer(server, s)

	server.Start()

	return nil
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// CheckUser returns whether the username and password are correct.
func (s *Server) CheckUser(ctx context.Context, req *pb.CheckUserRequest) (*pb.CheckUserResult, context.Context, error) {
	res := new(pb.CheckUserResult)

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
		log.Error().Msgf("Failed get users data: %s", err.Error())
	}

	res := make(map[string]string)
	for _, user := range users {
		res[user.Username] = user.Password
	}

	return res
}

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
