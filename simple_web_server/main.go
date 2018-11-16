package main

import (
	"flag"
	"log"
	"net/http"

	"./api"
	"./configuration"
	"./db"
)

const defaultConfigPath = "config.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to config file")
	flag.Parse()
	config, err := configuration.Init(configPath)

	db, err := db.Init(config)
	if err != nil {
		log.Panicln(err.Error())
	}
	defer db.Close()

	api := &api.API{
		DB: db,
	}
	r := api.InitRouter()

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Panicln(err.Error())
	}
}
