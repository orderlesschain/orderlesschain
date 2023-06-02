package main

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/serverconfig"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/nodeservice"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"net"
	"path/filepath"
)

func main() {
	if err := config.RemoveSavedData(); err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	cert, _ := filepath.Abs("./configs/cert.pem")
	key, _ := filepath.Abs("./configs/key.pem")
	transactionServer(cert, key)
}

func transactionServer(cert, key string) {
	lis, err := net.Listen("tcp", config.Config.TransactionServerPort)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	credential, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	opts := []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(serverconfig.KAEP),
		grpc.KeepaliveParams(serverconfig.KASP), grpc.Creds(credential)}
	grpcServer := grpc.NewServer(opts...)
	serviceInstance := nodeservice.NewTransactionService()
	protos.RegisterTransactionServiceServer(grpcServer, serviceInstance)
	go serviceInstance.LoadOtherNodesPublicKeys()
	logger.InfoLogger.Println("Transaction Server is listening on port", config.Config.TransactionServerPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
}
