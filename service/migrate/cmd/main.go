package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/MamangRust/microservice-ecommerce-pkg/database"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
)

const (
	dialect = "pgx"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "", "directory with migration files (default: per bounded context)")
)

// boundedContext groups the services that own one PostgreSQL instance. The
// order follows the dependency chain identity → merchant → catalog → sales →
// experience → email (each context validates cross-context references in the
// application, never via a cross-instance foreign key).
var boundedContexts = []struct {
	cluster  string
	services []string
}{
	{database.IdentityCluster, []string{"auth", "user", "role"}},
	{database.MerchantCluster, []string{"merchant", "merchant_award", "merchant_business", "merchant_detail", "merchant_policy"}},
	{database.CatalogCluster, []string{"category", "product"}},
	{database.SalesCluster, []string{"order", "order_item", "transaction"}},
	{database.ExperienceCluster, []string{"cart", "review", "review_detail", "slider", "shipping_address", "banner"}},
	{database.EmailCluster, []string{"email"}},
}

func main() {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := flags.Args()
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	command := args[0]
	extraArgs := args[1:]

	err := dotenv.Viper()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	if *dir != "" {
		// Explicit dir: delegate to goose directly against the base database.
		runOne(command, "", *dir, nil, extraArgs)
		return
	}

	for _, ctx := range boundedContexts {
		matches := collectMigrations(ctx.services)
		if len(matches) == 0 {
			log.Printf("Skipping %s: no migration files", ctx.cluster)
			continue
		}
		runOne(command, ctx.cluster, "", matches, extraArgs)
	}
}

// runOne applies one goose command to one target. When cluster is empty the
// legacy single-database DB_* keys are used — only reachable through the -dir
// flag, which nothing passes today. Otherwise the cluster prefix selects the
// instance: one PostgreSQL per bounded context, each behind its own PgBouncer.
// When explicitDir is set the files in matches are ignored.
func runOne(command, cluster, explicitDir string, matches, extraArgs []string) {
	if cluster == "" {
		cluster = "DB"
	}

	host := pick(cluster+"_HOST", "DB_HOST")
	port := pick(cluster+"_PORT", "DB_PORT")
	user := pick(cluster+"_USERNAME", "DB_USERNAME")
	dbName := pick(cluster+"_NAME", "DB_NAME")
	password := pick(cluster+"_PASSWORD", "DB_PASSWORD")

	// Same reasoning as database.NewClientWithPrefix: a missing prefixed key must
	// abort rather than let goose migrate whichever instance the empty host
	// resolves to.
	if host == "" || port == "" || dbName == "" {
		log.Fatalf("Incomplete database configuration for %s (host=%q port=%q dbname=%q)", cluster, host, port, dbName)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbName, password,
	)

	db, err := goose.OpenDBWithDriver(dialect, connStr)
	if err != nil {
		log.Fatalf("Error opening database %s (%s): %v", dbName, cluster, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Error closing database %s: %v", dbName, err)
		}
	}()

	if explicitDir != "" {
		if err := goose.RunContext(context.Background(), command, db, explicitDir, extraArgs...); err != nil {
			log.Fatalf("Migration failed on %s: %v", dbName, err)
		}
		return
	}

	// goose requires a plain directory, so stage this context's files into a
	// temp dir. Files are already sorted by timestamp; within a context the
	// original chronology is preserved.
	tmp, err := os.MkdirTemp("", "ec-migrations-")
	if err != nil {
		log.Fatalf("Failed to create temp migrations dir: %v", err)
	}
	defer os.RemoveAll(tmp)
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", m, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, filepath.Base(m)), data, 0o644); err != nil {
			log.Fatalf("Failed to stage %s: %v", filepath.Base(m), err)
		}
	}

	log.Printf("Migrating %s [%s] (%d files)", dbName, cluster, len(matches))
	if err := goose.RunContext(context.Background(), command, db, tmp, extraArgs...); err != nil {
		log.Fatalf("Migration failed on %s (%s): %v", dbName, cluster, err)
	}
}

// collectMigrations returns the timestamp-sorted migration files of the given
// services. Missing directories are skipped.
func collectMigrations(services []string) []string {
	var matches []string
	for _, svc := range services {
		found, err := filepath.Glob(filepath.Join("service", svc, "database", "migration", "*.sql"))
		if err != nil {
			log.Fatalf("Failed to glob migrations for %s: %v", svc, err)
		}
		matches = append(matches, found...)
	}
	sort.Strings(matches)
	return matches
}

// pick returns the prefixed key when set, otherwise the base key.
func pick(prefixed, base string) string {
	if v := viper.GetString(prefixed); v != "" {
		return v
	}
	return viper.GetString(base)
}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: migrate COMMAND
Examples:
    migrate status
    migrate up
`

	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations`
)
