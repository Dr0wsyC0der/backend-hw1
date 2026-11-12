package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)


func main() {
	timeout := flag.Int("t", 15, "request timeout in seconds")
	flag.IntVar(timeout, "timeout", 15, "request timeout in seconds")
	flag.Parse()
	urls := flag.Args()

	if len(urls) == 0 {
		fmt.Println("No URLs provided")
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(*timeout)*time.Second,
	)
	defer cancel()

	results := make(chan Result, len(urls))

	for _, url := range urls {
		go fetch(ctx, url, results)
	}

	errorsCount := 0
	for errorsCount < len(urls) {
		select {
		case result := <-results:
			if result.Err != nil {
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					os.Exit(228)
				}
				errorsCount++
				if errorsCount == len(urls) {
					fmt.Println("All requests failed")
					os.Exit(1)
				}
				continue
			}

			fmt.Printf("%s %s\n", result.Resp.Proto, result.Resp.Status)
			for key, values := range result.Resp.Header {
				for _, value := range values {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
			fmt.Printf("Response body: %s\n", string(result.Body))
			cancel()
			return

		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				os.Exit(228)
			}
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
		}
	}

	if errorsCount == len(urls) {
		fmt.Println("All requests failed")
		os.Exit(1)
	}
}
