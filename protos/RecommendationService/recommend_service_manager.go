package RecommendationService

import (
	"errors"
	"log"
	"rickycorte/maki/conf"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var makiCfg *conf.Configuration

var currentConnection *grpc.ClientConn
var currentClient *RecommendationServiceClient

func openNewConnection() (*grpc.ClientConn, *RecommendationServiceClient) {
	conn, err := grpc.Dial(
		makiCfg.RecommendationServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Could not connect to gRPC: %v", err)
		return nil, nil
	}

	client := NewRecommendationServiceClient(conn)

	return conn, &client
}

func GetRecommendationServiceClient() (RecommendationServiceClient, error) {

	var err error = nil

	if currentConnection == nil {
		currentConnection, currentClient = openNewConnection()
	}
	if currentClient == nil {
		err = errors.New("unable to create gRPC client")
	}

	return *currentClient, err
}

func SetMakiConfig(cfg *conf.Configuration) {
	makiCfg = cfg
}
