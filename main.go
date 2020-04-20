package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

func readPostgresSettings(useEnvFile bool) (dbName, dbUser, dbPassword string) {
	if useEnvFile {
		godotenv.Load()
	}

	dbName = os.Getenv("POSTGRES_DB")
	dbUser = os.Getenv("POSTGRES_USER")
	dbPassword = os.Getenv("POSTGRES_PASSWORD")

	if dbName == "" {
		dbName = "postgres"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPassword == "" {
		dbPassword = "ito"
	}

	return
}

func main() {
	var (
		port       string
		dbHost     string
		useEnvFile bool
	)

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Value:       "8080",
				Usage:       "Port for the server to run on",
				Destination: &port,
			},
			&cli.StringFlag{
				Name:        "dbhost",
				Value:       "127.0.0.1",
				Usage:       "The Postgres host to be used",
				Destination: &dbHost,
			},
			&cli.BoolFlag{
				Name:        "env",
				Usage:       "Set to true to read from environment variable file",
				Destination: &useEnvFile,
			},
		},
		Action: func(ctx *cli.Context) error {
			dbName, dbUser, dbPassword := readPostgresSettings(useEnvFile)
			dbConnection, err := NewDBConnection(dbHost, dbUser, dbPassword, dbName)
			if err != nil {
				return err
			}
			return GetRouter(port, dbConnection).Run(fmt.Sprintf(":%s", port))
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
