package main

import (
	"flag"
	"gitlab.lrz.de/orderless/orderlesschain/internal/benchmark"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/serverconfig"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"net"
)

var (
	benchmarkName = flag.String("benchmark", "", "The benchmark configuration")
	isCoordinator = flag.Bool("coordinator", false, "Is this a coordinator for benchmarking")
)

func main() {
	if err := config.RemoveSavedData(); err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	flag.Parse()
	if *isCoordinator {
		if len(*benchmarkName) == 0 {
			logger.FatalLogger.Fatalln("benchmark must be provided")
		}
		be := benchmark.NewCoordinator(*benchmarkName)
		be.CoordinateBenchmark()
	} else {
		lis, err := net.Listen("tcp", config.Config.TransactionServerPort)
		if err != nil {
			logger.FatalLogger.Fatalln(err)
		}
		opts := []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(serverconfig.KAEP), grpc.KeepaliveParams(serverconfig.KASP)}
		grpcServer := grpc.NewServer(opts...)
		protos.RegisterBenchmarkServiceServer(grpcServer, benchmark.NewExecutorServer())
		logger.InfoLogger.Println("Client Server is listening on port", config.Config.TransactionServerPort)
		err = grpcServer.Serve(lis)
		if err != nil {
			logger.FatalLogger.Fatalln(err)
		}
	}
}
