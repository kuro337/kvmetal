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

	wg.Add(1)
	go func() {
		defer wg.Done()
		cli.Evaluate(ctx, &wg)
	}()

	go func() {
		wg.Wait()
		stop()
	}()

	<-ctx.Done()
	log.Println("All operations completed, exiting.")
}

/*


go run main.go --launch-vm=control  --preset=kubecontrol --mem=4096 --cpu=2
go run main.go --launch-vm=worker   --preset=kubeworker  --mem=4096 --cpu=2

go run main.go --join=control,worker


*/
