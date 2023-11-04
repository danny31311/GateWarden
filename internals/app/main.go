package app

import (
	"GateWarden/api"
	"GateWarden/api/middleware"
	"GateWarden/internals/app/db"
	"GateWarden/internals/app/handlers"
	"GateWarden/internals/app/processors"
	"GateWarden/internals/cfg"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"log"
	"net/http"
	"time"
)

type Server struct {
	config cfg.Cfg
	ctx    context.Context
	srv    *http.Server
	db     *pgxpool.Pool
}

func NewServer(config cfg.Cfg, ctx context.Context) *Server {
	server := new(Server)
	server.ctx = ctx
	server.config = config
	return server
}

func (server *Server) Serve() {
	log.Println("Starting server")
	var err error
	server.db, err = pgxpool.Connect(server.ctx, server.config.GetDbString())
	if err != nil {
		log.Fatal(err)
	}
	carsStorage := db.NewCarStorage(server.db)
	usersStorage := db.NewUsersStorage(server.db)

	carsProcessor := processors.NewCarsProcessor(carsStorage)
	usersProcessor := processors.NewUsersProcessor(usersStorage)

	userHandler := handlers.NewUsersHandler(usersProcessor)
	carsHandler := handlers.NewCarsHandler(carsProcessor)

	routes := api.CreateRoutes(userHandler, carsHandler)
	routes.Use(middleware.RequestLog)

	server.srv = &http.Server{
		Addr:    ":" + server.config.Port,
		Handler: routes,
	}

	log.Println("Server started")
	err = server.srv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
	return
}

func (server *Server) ShutDown() {
	log.Printf("Server stopped")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.db.Close()
	defer func() {
		cancel()
	}()
	var err error
	if err = server.srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server shutdown failed : %v", err)
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

}
