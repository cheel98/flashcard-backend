package main

import (
	"context"
	"github.com/cheel98/flashcard-backend/internal/app"
	"log"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		app.Module,
		fx.Invoke(func(lc fx.Lifecycle, server *app.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Println("Starting flashcard backend server...")
					return server.Start()
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Stopping flashcard backend server...")
					return server.Stop()
				},
			})
		}),
	)

	app.Run()
}
