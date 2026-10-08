package main

import (
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/apps"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "merchant_policy-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50068,
		DBCluster:      "DB_MERCHANT",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
