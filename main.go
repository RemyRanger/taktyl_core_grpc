package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"

	"github.com/RemyRanger/taktyl_core_grpc/src/gateway"
	"github.com/RemyRanger/taktyl_core_grpc/src/seed"
	"github.com/RemyRanger/taktyl_core_grpc/src/server"
	"github.com/joho/godotenv"

	// Proto Injects
	pbEvent "github.com/RemyRanger/taktyl_core_grpc/src/proto/event"
	pbUser "github.com/RemyRanger/taktyl_core_grpc/src/proto/user"

	// Static files
	_ "github.com/RemyRanger/taktyl_core_grpc/statik"
)

func main() {

	// Setting env values
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	} else {
		fmt.Println("We are getting the env values")
	}

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := grpc.NewServer(
	// TODO: Replace with your own certificate!
	//grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
	)

	backend := server.New().Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	pbUser.RegisterUserServiceServer(s, backend)
	pbEvent.RegisterEventServiceServer(s, backend)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	seed.Load(backend.DB)

	// Serve gRPC Server
	log.Info("Serving gRPC on https://", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	err = gateway.Run("dns:///" + addr)
	log.Fatalln(err)
}
