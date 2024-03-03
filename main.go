package main

import (
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

var (
	LISTEN   = os.Getenv("LISTEN")
	STORAGE  = os.Getenv("STORAGE")
	ROOT_KEY = os.Getenv("ROOT_KEY")
)

func init() {
	if LISTEN == "" {
		LISTEN = ":24303" //default listen addr
	}
	if STORAGE == "" {
		STORAGE = "./data" //default storage dir
	}

	var ok bool
	if ROOT_KEY == "" {
		ROOT_KEY, ok = MetaOf("/").WriteKey()
		if !ok {
			ROOT_KEY = uuid.NewString()

			MetaOf("/").SetWriteKey(ROOT_KEY)
			log.Println("generated root key:", ROOT_KEY)
		}
	}
}

func main() {
	log.Println("listening at", LISTEN)
	http.ListenAndServe(LISTEN, newServer())
}
