package main

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-order/apps"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "order-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50057,
		DBCluster:      "DB_SALES",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
