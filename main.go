package main

import (
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	pingCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_ping_request_count",
		Help: "No of request handled by Ping handler",
	})
)

func extralogger(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("query headers %+v", r.Header)
		next.ServeHTTP(w, r)
	})
}

func logger(next http.Handler) http.HandlerFunc {
	return extralogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("query received %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}))
}

func handlePing() http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
    	pingCounter.Inc()
	    log.Printf("pong")
	    w.Write([]byte("PONG"))
    }
}

func handleIndex() http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        writer.Write([]byte("Welcome"))
    }
}

func main() {
	recordMetrics()

	//http.Handle("/ping", logger(extralogger(http.HandlerFunc(ping))))
	//http.Handle("/metrics", logger(promhttp.Handler()))

	router := httprouter.New()
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			// Set CORS headers
			header := w.Header()
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", "*")
		}

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	router.HandlerFunc("GET", "/", handleIndex())
	router.HandlerFunc("GET", "/ping", logger(extralogger(http.HandlerFunc(handlePing()))))
	router.HandlerFunc("GET", "/metrics", logger(promhttp.Handler()))

	http.ListenAndServe(":8090", router)
}
