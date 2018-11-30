package main

import (
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
	l := logger.InitLogger()
	defer l.Sync()

	prometheus.MustRegister(metrics.AccessHits)

	dm := database.InitDatabaseManager("postgres@postgres:5432", "ketnipz")
	defer dm.Close()

	sm := session.ConnectSessionManager()
	defer sm.Close()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc(
		"/session", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.SessionHandler(dm, sm), sm))))),
	)
	http.HandleFunc(
		"/profile", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.ProfileHandler(dm, sm), sm))))),
	)
	http.HandleFunc(
		"/profile/avatar", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.AvatarHandler(dm), sm))))),
	)
	http.HandleFunc(
		"/scoreboard", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(handlers.ScoreboardHandler(dm))))),
	)

	// swag init -g handlers/api.go
	http.HandleFunc("/api/docs/", httpSwagger.WrapHandler)

	http.HandleFunc(
		"/static/", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(filesystem.StaticHandler)))),
	)

	logger.Info("starting server at: ", 8080)
	logger.Panic(http.ListenAndServe(":8080", nil))
}
