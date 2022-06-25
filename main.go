package main

import (
        "net/http"
        "time"
        "log"

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
                Name: "ping_request_count",
                Help: "No of request handled by Ping handler",
        })
)

func ping(w http.ResponseWriter, req *http.Request) {
   pingCounter.Inc()
   log.Printf("pong")
}


func logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("query %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
func main() {
        recordMetrics()

        http.HandleFunc("/ping", logger(ping))
        http.Handle("/metrics", logger(promhttp.Handler()))
        http.ListenAndServe(":8090", nil)
}
