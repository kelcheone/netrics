package main

import (
	"log"
	"net/http"

	sniffer "github.com/kelcheone/netrics/cmd"
	prometheusdb "github.com/kelcheone/netrics/cmd/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := Start(); err != nil {
		log.Fatalln(err)
	}
}

func Start() error {
	go func() {
		err := sniffer.Sniff(prometheusdb.LogUsage)
		if err != nil {
			log.Println("Error monitoring the network: ", err)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		return err
	}
	return nil
}
