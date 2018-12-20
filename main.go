package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/middleware"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"

	_ "api/docs"
	"api/filesystem"
	"api/handlers"
	"api/metrics"
)

func main() {
	dbConnStr := flag.String("db_connstr", "postgres@localhost:5432", "postgresql connection string")
	dbName := flag.String("db_name", "postgres", "database name")
	authConnStr := flag.String("auth_connstr", "localhost:8081", "auth-service connection string")
	flag.Parse()

	l := logger.InitLogger()
	defer func() {
		err := l.Sync()
		if err != nil {
			logger.Errorf("error while syncing log data: %v", err)
		}
	}()

	prometheus.MustRegister(metrics.AccessHits)

	dm := database.InitDatabaseManager(*dbConnStr, *dbName)
	defer dm.Close()

	sm := session.ConnectSessionManager(*authConnStr)
	defer sm.Close()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc(
		"/session",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.SessionHandler(dm, sm), sm))))),
	)
	http.HandleFunc(
		"/profile",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.ProfileHandler(dm, sm), sm))))),
	)
	http.HandleFunc(
		"/profile/avatar",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.AvatarHandler(dm), sm))))),
	)
	http.HandleFunc(
		"/profile/skin",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.SkinHandler(dm), sm))))),
	)
	http.HandleFunc(
		"/profile/check",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(handlers.CheckAvailabilityHandler(dm))))),
	)
	http.HandleFunc(
		"/scoreboard",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(handlers.ScoreboardHandler(dm))))),
	)

	// swag init -g handlers/api.go
	http.HandleFunc("/api/docs/", httpSwagger.WrapHandler)

	stm := filesystem.NewStaticManager("/static/", "static")

	http.HandleFunc(
		"/static/",
		middleware.RecoverMiddleware(metrics.CountHitsMiddleware(middleware.AccessLogMiddleware(
			middleware.CORSMiddleware(stm)))),
	)

	logger.Info("starting server at: ", 8080)
	logger.Panic(http.ListenAndServe(":8080", nil))
}
