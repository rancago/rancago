package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rancago/framework/internal/bootstrap"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println(`
   ______       _       __
  / ____/___ _/ |     / /__  ____ _____
 / / __/ __  /| | /| / / _ \/ __  / __ \
/ /_/ / /_/ / | |/ |/ /  __/ /_/ / /_/ /
\____/\__,_/  |__/|__/\___/\__, /\____/
                          /____/
`)
	app := bootstrap.New()
	log.Println("[rancago] Bootstrapping application (hexagonal architecture)...")
	app.RegisterCore()
	app.Boot()
	log.Println("[rancago] Application booted successfully.")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := app.StartHTTP(ctx); err != nil {
			log.Printf("[rancago] HTTP server stopped: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.StartGRPC(ctx); err != nil {
			log.Printf("[rancago] gRPC server stopped: %v", err)
		}
	}()
	log.Printf("[rancago] HTTP: :%d  |  gRPC: :%d  |  WebSocket: /ws",
		app.Config.Server.HTTPPort, app.Config.Server.GRPCPort)
	<-ctx.Done()
	log.Println("[rancago] Shutdown signal received, draining...")
	wg.Wait()
	log.Println("[rancago] Goodbye!")
	os.Exit(0)
}
