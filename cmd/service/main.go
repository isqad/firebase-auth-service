package main

import (
	"log"
	"os"

	"github.com/isqad/firebase-auth-service/pkg/service"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "firebase-auth-service",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "listen-address", Aliases: []string{"l"}, Value: ":50053"},
		},
		Action: startServer,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func startServer(ctx *cli.Context) error {
	server, err := service.NewAPIServer()
	if err != nil {
		return err
	}

	return server.Start(ctx)
}
