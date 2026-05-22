package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/seeders"
	"github.com/joho/godotenv"
	"github.com/kristijorgji/goseeder"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: .env not loaded (%v) — continuing with environment variables\n", err)
	}

	fresh := flag.Bool("fresh", false, "truncate all seeded tables before running (DESTRUCTIVE)")
	only := flag.String("only", "", "comma-separated seeder names to run, e.g. SeedUsers,SeedTransactions")
	flag.Parse()

	db := config.Connect()
	defer db.Close()

	if *fresh {
		config.Truncate(db)
	}

	goseeder.Register(seeders.SeedUsers)
	goseeder.Register(seeders.SeedFinancialProfiles)
	goseeder.Register(seeders.SeedTransactions)
	goseeder.Register(seeders.SeedBudgets)
	goseeder.Register(seeders.SeedGoals)
	goseeder.Register(seeders.SeedAiLogs)

	var opts []goseeder.ConfigOption
	if *only != "" {
		opts = append(opts, goseeder.ForSpecificSeeds(strings.Split(*only, ",")))
	}

	if err := goseeder.Execute(db, opts...); err != nil {
		fmt.Fprintf(os.Stderr, "seeder: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSeeding complete!")
}
