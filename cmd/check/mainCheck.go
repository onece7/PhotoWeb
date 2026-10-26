package main

import (
	"fmt"
	"log"
	"novel2comic/backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var dbName, schema, user string
	db.Raw("SELECT current_database()").Scan(&dbName)
	db.Raw("SELECT current_schema()").Scan(&schema)
	db.Raw("SELECT current_user").Scan(&user)
	fmt.Printf("连的库=%s schema=%s 用户=%s\n", dbName, schema, user)

}
