package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	var (
		port       string
		dbHost     string
		dbUser     string
		dbPassword string
		dbName     string
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
			&cli.StringFlag{
				Name:        "dbuser",
				Value:       "",
				Usage:       "The Postgres user to be used",
				Destination: &dbUser,
			},
			&cli.StringFlag{
				Name:        "dbpw",
				Value:       "",
				Usage:       "The password to be used when connecting to Postgres",
				Destination: &dbPassword,
			},
			&cli.StringFlag{
				Name:        "dbname",
				Value:       "",
				Usage:       "The name of the Postgres database to be used",
				Destination: &dbName,
			},
		},
		Action: func(ctx *cli.Context) error {
			dbConnection, err := NewDBConnection(dbHost, dbUser, dbPassword, dbName)
			if err != nil {
				return err
			}
			if err := StartServer(port, dbConnection); err != nil {
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
