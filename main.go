package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/locke105/magicbox/proxy"
)

func main() {
	url, err := url.Parse("http://localhost:8942")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting\n")
	proxy := proxy.NewRecordingProxy(url)
	http.ListenAndServe(":3128", proxy)
}
