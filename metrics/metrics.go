package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/wrappers"
)

const (
	PrometheusNamespace = "api_service"
)

var (
	AccessHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: PrometheusNamespace,
		Name:      "hits_by_http_status",
		Help:      "Total hits ordered by http response statuses",
	},
		[]string{"http_status", "path", "method"},
	)
)

func MetricsHitsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &wrappers.StatusWrapperForResponseWriter{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		next.ServeHTTP(ww, r)

		AccessHits.With(prometheus.Labels{
			"http_status": strconv.Itoa(ww.Status),
			"path":        r.URL.Path,
			"method":      r.Method,
		}).Inc()
	})
}
