package main

import (
	"context"
	"io"
	"net/http"
)

func fetch(ctx context.Context, url string, results chan<- Result) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		results <- Result{Err: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		results <- Result{Err: err}
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		results <- Result{Err: err}
		return
	}

	results <- Result{Resp: resp, Body: body}

}
