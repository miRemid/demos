package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/miRemid/demos/grpc/api/v1/cqless"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type simple struct {
	cqless.UnimplementedCQLessServiceServer
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	lis, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	cqless.RegisterCQLessServiceServer(grpcServer, simple{})

	err = cqless.RegisterCQLessServiceHandlerFromEndpoint(ctx, mux, "localhost:9090", opts)
	if err != nil {
		panic(err)
	}
	log.Println("start http gateway")
	go http.ListenAndServe(":8081", mux)
	log.Println("start grpc server")
	grpcServer.Serve(lis)
}
