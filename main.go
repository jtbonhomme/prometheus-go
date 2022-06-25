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

func handlerFuncAdapter(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
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

type Route struct {
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

func contact() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Contact"))
	}
}

func home() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Home"))
	}
}

var Routes = []Route{
    Route{"GET", "/home", home()},
    Route{"GET", "/contact", logger(contact())},
}

func setRoutes(router *httprouter.Router) {
    for _, route := range Routes {
        router.HandlerFunc(route.Method, route.Pattern, route.HandlerFunc)
    }
}

func main() {
	recordMetrics()

	router := httprouter.New()
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			// Set CORS headers
			header := w.Header()
			header.Set("Access-Control-Allow-Heders", "Access-Control-Expose-Headers, Content-Type, Content-Compression, X-Requested-With")
			header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			header.Set("Access-Control-Allow-Origin", "*")
		}

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	router.HandlerFunc("GET", "/", logger(handleIndex()))
	router.HandlerFunc("GET", "/ping", logger(extralogger(http.HandlerFunc(handlePing()))))
	router.HandlerFunc("GET", "/metrics", handlerFuncAdapter(promhttp.Handler()))

    setRoutes(router)

	http.ListenAndServe(":8090", router)
}
