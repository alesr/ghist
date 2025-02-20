package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/alesr/ghist/internal/ghclient"
	"github.com/alesr/ghist/internal/repository"
	"github.com/alesr/ghist/internal/service"
	"github.com/wiselead-ai/httpclient"
)

func main() {
	ghClient := ghclient.New(httpclient.New())

	sqliteRepo, err := repository.NewSQLite()
	if err != nil {
		log.Fatal(err)
	}

	svc := service.New(slog.Default(), ghClient, sqliteRepo)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	diffs, err := svc.Run(ctx, "alesr")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range diffs {
		fmt.Println(d.String())
	}
}
