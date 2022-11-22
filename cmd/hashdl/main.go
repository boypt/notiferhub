package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	webListen string
	dlCache   *cache.Cache
)

func dlSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	uuid := r.PostFormValue("uuid")
	path := r.PostFormValue("path")
	validtime := r.PostFormValue("validtime")
	log.Println("dlset", uuid, path)

	if uuid == "" || path == "" {
		http.NotFound(w, r)
		return
	}

	expire := cache.DefaultExpiration
	if validtime != "" {
		if e, err := time.ParseDuration(validtime); err == nil {
			expire = e
		}
	}

	dlCache.Set(uuid, path, expire)
	fmt.Fprintf(w, "OK")
}

func dlReq(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.RawQuery
	fwdAddress := r.Header.Get("X-Forwarded-For") // capitalisation doesn't matter

	if p, ok := dlCache.Get(uuid); ok {
		if err := dlCache.Replace(uuid, p, time.Hour); err != nil {
			log.Println("dlReq", uuid, "replace failed", err)
		}

		fp := p.(string)
		log.Printf("dlreq from [%s] got [%s] : [%s]", fwdAddress, uuid, fp)
		w.Header().Set("X-Accel-Redirect", path.Join("/protected", fp))
		return
	}

	log.Println("dlreq from ", fwdAddress, "not found", uuid)
	http.NotFound(w, r)
}

func main() {

	flag.StringVar(&webListen, "webport", "127.0.0.1:5333", "web server listen")
	flag.Parse()

	dlCache = cache.New(72*time.Hour, 30*time.Minute)

	http.HandleFunc("/", dlReq)
	http.HandleFunc("/dlset", dlSet)

	fmt.Println("Starting hash dl server at ", webListen)
	if err := http.ListenAndServe(webListen, nil); err != nil {
		log.Fatal(err)
	}
}
