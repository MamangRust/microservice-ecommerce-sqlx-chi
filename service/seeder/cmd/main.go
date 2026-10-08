package main

import (
	"fmt"
	"os"
	"os/exec"
)

// domains lists every per-service seeder in dependency order. The order is the
// seed contract from PLANNING_MICROSERVICE.md §6.3: identity → catalog →
// merchant → product → order → transaction → remaining content. Each entry is
// run via `go run ./service/<domain>/cmd/seeder` so the seed logic lives in the
// service that owns the tables (F2). Run this orchestrator from the repo root;
// the per-service seeders read the same DB_* env vars the services use.
var domains = []string{
	"user",
	"role",
	"merchant",
	"merchant_detail",
	"merchant_award",
	"merchant_policy",
	"merchant_business",
	"category",
	"product",
	"order",
	"shipping_address",
	"review",
	"transaction",
	"slider",
	"banner",
}

func main() {
	for _, d := range domains {
		fmt.Printf("== seeding %s ==\n", d)
		cmd := exec.Command("go", "run", "./service/"+d+"/cmd/seeder")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "seeder %s failed: %v\n", d, err)
			os.Exit(1)
		}
	}
	fmt.Println("All domains seeded successfully")
}
