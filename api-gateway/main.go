package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}

func main() {
	userService := newProxy("http://user-service:3001")
	menuService := newProxy("http://menu-service:3002")
	orderService := newProxy("http://order-service:3003")

	mux := http.NewServeMux()

	mux.Handle("/users/", userService)
	mux.Handle("/users", userService)

	mux.Handle("/menus/", menuService)
	mux.Handle("/menus", menuService)

	mux.Handle("/orders/", orderService)
	mux.Handle("/orders", orderService)

	log.Println("API Gateway running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
