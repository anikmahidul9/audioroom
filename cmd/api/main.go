package main

import (
	"log"

	"github.com/anikmahidul9/voiceroom/internal/db"
	"github.com/anikmahidul9/voiceroom/internal/env"
	"github.com/anikmahidul9/voiceroom/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config{
		addr: env.GetString("ADDR", ":8000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:env.GetString("ENV","development"),
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	log.Println("Database connection established")
	store := store.NewPostgresStorage(db)

	app := application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	if err := app.run(mux); err != nil {
		panic(err)
	}
}
