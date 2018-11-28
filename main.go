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

	db := database.InitDB("postgres@postgres:5432", "ketnipz")
	defer db.Close()

	sm := session.ConnectSessionManager()
	defer sm.Close()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc(
		"/session", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.SessionHandler(sm), sm))))),
	)
	http.HandleFunc(
		"/profile", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.ProfileHandler(sm), sm))))),
	)
	http.HandleFunc(
		"/profile/avatar", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(middleware.SessionMiddleware(handlers.AvatarHandler, sm))))),
	)
	http.HandleFunc(
		"/scoreboard", 
		middleware.RecoverMiddleware(metrics.MetricsHitsMiddleware(middleware.AccessLogMiddleware(
		middleware.CORSMiddleware(handlers.ScoreboardHandler)))),
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
