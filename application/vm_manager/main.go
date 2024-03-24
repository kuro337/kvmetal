package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"kvmgo/cli"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Kick off your asynchronous operation
	wg.Add(1) // Increment the WaitGroup counter
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		cli.Evaluate(ctx, &wg)
	}()

	// Wait for the operation to complete
	go func() {
		wg.Wait()
		stop() // Cancel the context once all operations have completed
	}()

	<-ctx.Done()
	log.Println("All operations completed, exiting.")
}
