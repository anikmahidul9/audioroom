package main

import "github.com/anikmahidul9/voiceroom/internal/env"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	app := application{
		config: cfg,
	}

	mux := app.mount()
	if err := app.run(mux); err != nil {
		panic(err)
	}
}
