package main

import (
	"fmt"
	"log"

	"private/ibms/internal/config"
	"private/ibms/internal/route"
	"private/ibms/internal/service"
	"private/ibms/internal/store"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := store.Open(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// 自底向上串联三层：store → service → route。
	st := store.New(db)
	svc := service.New(st)
	r := route.New(svc)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
