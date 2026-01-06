package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anton2920/gofa/bools"

	"golang.org/x/net/webdav"
)

func Request2String(r *http.Request) string {
	return fmt.Sprintf("%s ")
}

func main() {
	addr := flag.String("addr", "0.0.0.0:8080", "address:port to bind on")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	levels := [...]string{"INFO", "ERROR"}
	handler := webdav.Handler{
		Prefix:     "",
		FileSystem: webdav.Dir(wd),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, ierr error) {
			level := levels[bools.ToInt(ierr != nil)]
			log.Printf("%5s [%21s] %8s %s -> %v", level, r.RemoteAddr, r.Method, r.URL.Path, ierr)
		},
	}

	log.Printf(" INFO Listening on %s...", *addr)
	if err := http.ListenAndServe(*addr, &handler); (err != nil) && (err != http.ErrServerClosed) {
		log.Fatalf("FATAL Failed to listen and serve on %s: %v", addr, err)
	}
}
