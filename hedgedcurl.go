package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func readBody(r *http.Response) ([]byte, error) {
	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

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
				errorsCount++
				if errorsCount == len(urls) {
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						os.Exit(228)
					}
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
			body, err := readBody(result.Resp)
			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Response body: %s\n", string(body))
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
