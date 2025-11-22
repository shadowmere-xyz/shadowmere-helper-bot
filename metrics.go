package main

import (
	"fmt"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	errorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shadowmere_helper_bot_errors_total",
		Help: "The total number of errors found",
	})
	additionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shadowmere_helper_bot_successful_additions_total",
		Help: "The total number of proxies successfully added",
	})
	additionErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shadowmere_helper_bot_addition_errors_total",
		Help: "The total number of proxies successfully added",
	})
)

func startMetricsServer(port int) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting metrics server on port %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
