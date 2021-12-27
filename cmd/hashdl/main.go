package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	webListen string
	dlCache   *cache.Cache
)

func dlSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Only POST method is supported")
		return
	}

	uuid := r.PostFormValue("uuid")
	path := r.PostFormValue("path")
	log.Println("dlset", uuid, path)

	if uuid == "" || path == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "No value")
		return
	}

	dlCache.Set(uuid, path, cache.DefaultExpiration)
	fmt.Fprintf(w, "OK")
}

func dlReq(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimPrefix(r.URL.Path, "/dlreq/")
	fwdAddress := r.Header.Get("X-Forwarded-For") // capitalisation doesn't matter

	if p, ok := dlCache.Get(uuid); ok {
		fp := p.(string)
		log.Printf("dlreq from [%s] got [%s] : [%s]", fwdAddress, uuid, fp)
		w.Header().Set("X-Accel-Redirect", path.Join("/protected", fp))
		return
	}

	log.Println("dlreq from ", fwdAddress, "not found", uuid)
	w.WriteHeader(404)
	fmt.Fprintf(w, "Not found")
}

func main() {

	flag.StringVar(&webListen, "webport", "127.0.0.1:5333", "web server listen")
	flag.Parse()

	dlCache = cache.New(30*time.Minute, 10*time.Minute)

	http.HandleFunc("/dlreq/", dlReq)
	http.HandleFunc("/dlset", dlSet)

	fmt.Println("Starting hash dl server at ", webListen)
	if err := http.ListenAndServe(webListen, nil); err != nil {
		log.Fatal(err)
	}
}
