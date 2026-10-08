package main

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-category/apps"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "category-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50054,
		DBCluster:      "DB_CATALOG",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
