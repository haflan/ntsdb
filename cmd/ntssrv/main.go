package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	ndb "github.com/haflan/ntsdb"
	"github.com/pkg/browser"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: ntssrv <path> [addr]")
		return
	}
	dbpath := args[0]
	addr := ":2015"
	if len(args) > 1 {
		addr = args[1]
	}

	ntsdb := ndb.OpenNTSDB(dbpath)
	defer ntsdb.Close()
	srv := NewNoteServer(ntsdb)
	log.Printf("Serving '%s' at %s", dbpath, addr)
	go func() {
		log.Fatal(http.ListenAndServe(addr, srv))
	}()

	// Open browser
	if strings.HasPrefix(addr, ":") {
		addr = "localhost" + addr
	}
	browser.OpenURL(addr)
	select {}
}
