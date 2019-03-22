package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/dsociative/incrementer/api"
	"github.com/dsociative/incrementer/db"
)

var (
	redisAddr      = flag.String("redis", "localhost:6379", "redis addr")
	bind           = flag.String("bind", ":8080", "bind addr")
	provisionMax   = flag.Int("max", 1000, "provision maximum value")
	provisionStep  = flag.Int("step", 1, "provision step value")
	provisionValue = flag.Int("value", 0, "provision initial value")
)

func main() {
	flag.Parse()
	db := db.NewRedis(*redisAddr)
	err := db.Provision(*provisionMax, *provisionStep, *provisionValue)
	if err != nil {
		log.Printf("Provision DB failed: %s", err)
	} else {
		log.Println("Provision DB OK")
	}
	twirpHandler := api.NewIncrementerServer(api.NewAPI(db), nil)
	http.ListenAndServe(*bind, twirpHandler)
}
