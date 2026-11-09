package main

import "net/http"

type Result struct {
	Resp *http.Response
	Err  error
}
