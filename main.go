package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")

	if dbName == "" || dbUser == "" || dbPassword == "" {
		panic("Error loading environment variables")
	}

	var (
		port   string
		dbHost string
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
		},
		Action: func(ctx *cli.Context) error {
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
