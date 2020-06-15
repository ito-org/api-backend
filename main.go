package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/openmined/tcn-psi/server"
)

func readPostgresSettings() (dbHost, dbName, dbUser, dbPassword string) {
	dbHost = os.Getenv("POSTGRES_HOST")
	dbName = os.Getenv("POSTGRES_DB")
	dbUser = os.Getenv("POSTGRES_USER")
	dbPassword = os.Getenv("POSTGRES_PASSWORD")

	if dbHost == "" {
		dbHost = "localhost"
	}
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
	var port string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Value:       "8080",
				Usage:       "Port for the server to run on",
				Destination: &port,
			},
		},
		Action: func(ctx *cli.Context) error {
			dbHost, dbName, dbUser, dbPassword := readPostgresSettings()
			dbConnection, err := NewDBConnection(dbHost, dbUser, dbPassword, dbName)
			if err != nil {
				return err
			}

			psicServer, err := server.CreateWithNewKey()
			if err != nil {
				return err
			}

			return GetRouter(port, dbConnection, psicServer).Run(fmt.Sprintf(":%s", port))
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
