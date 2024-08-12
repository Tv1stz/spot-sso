package app

import (
	"log/slog"
	grpcapp "ssov2/internal/app/grpc"
	"ssov2/internal/service/auth"
	mongoauth "ssov2/internal/storage/mongo/auth"
	"time"
)

type App struct {
	log  *slog.Logger
	port int
}

func New(
	log *slog.Logger,
	port int,
	mongoDBUri string,
) *App {

	storage, err := mongoauth.New(mongoDBUri)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, time.Hour, storage, storage)

	application := grpcapp.New(log, authService, port)
	application.MustRun()

	return &App{
		log:  log,
		port: port,
	}
}
