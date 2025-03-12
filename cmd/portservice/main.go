package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portservice/internal/adapters/secondary/memory"
	"portservice/internal/core"
	"portservice/internal/ports/out"
)

func main() {
	// Parse command line flags
	filePath := flag.String("file", "ports.json", "Path to the ports JSON file")
	flag.Parse()

	// Create repository and service
	repo := memory.NewPortRepository()
	service := core.NewPortService(repo)

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start processing in a goroutine
	errChan := make(chan error, 1)
	startTime := time.Now()
	go func() {
		errChan <- service.ProcessPortsFile(ctx, *filePath)
	}()

	// Wait for either completion or interruption
	var err error
	select {
	case err = <-errChan:
		if err == context.Canceled {
			log.Println("Processing was canceled")
		} else if err != nil {
			log.Printf("Error processing file: %v", err)
		} else {
			duration := time.Since(startTime)
			log.Printf("File processing completed successfully in %v", duration)

			// Display repository statistics
			if stats, ok := repo.(interface{ GetStatistics() out.RepositoryStats }); ok {
				repoStats := stats.GetStatistics()
				log.Printf("Repository statistics:")
				log.Printf("  - Total ports processed: %d", repoStats.TotalPorts)
				log.Printf("  - Total updates: %d", repoStats.TotalUpdates)
				log.Printf("  - Last update: %v", repoStats.LastUpdate)
				log.Printf("  - Average processing speed: %.2f ports/second",
					float64(repoStats.TotalPorts)/duration.Seconds())
			}
		}
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
		// Wait for processing to stop or timeout
		select {
		case err = <-errChan:
			log.Printf("Processing stopped with error: %v", err)
		case <-time.After(5 * time.Second):
			log.Println("Processing shutdown timed out")
		}
	}

	// Close repository
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCancel()

	if closeErr := repo.(interface{ Close(context.Context) error }).Close(closeCtx); closeErr != nil {
		log.Printf("Error closing repository: %v", closeErr)
	}

	if err != nil && err != context.Canceled {
		os.Exit(1)
	}
	log.Println("Service stopped")
}
