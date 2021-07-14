package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/um198/finalproject/cmd/app"
	"github.com/um198/finalproject/pkg/files"
	"github.com/um198/finalproject/pkg/security"
	"go.uber.org/dig"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	dns := "postgres://app:pass@localhost:5432/db"
	if err := execute(host, port, dns); err != nil {
		os.Exit(1)
	}
}

func execute(host string, port string, dns string) (err error) {
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter,
		func() (*pgxpool.Pool, error) {
			ctx, c := context.WithTimeout(context.Background(), time.Second*5)
			println("Server starts", c)
			return pgxpool.Connect(ctx, dns)
		},
		files.NewService,
		security.NewService,
		
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		return err
	}
	return container.Invoke(func(server *http.Server) error {
		return server.ListenAndServe()
	})
}
