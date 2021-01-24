package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var release = "dev" // set by build process
const updateDelay = 10 * time.Second

var (
	listenAddr = flag.String("web.listen-addr", ":9590", "Listening Address")
)

var (
	lastUpdate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bcg_last_update",
		Help: "Last bcg update",
	})
)

func main() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Printf("Usage for bcg-exporter (%s) https://github.com/natesales/bcg-exporter:\n", release)
		flag.PrintDefaults()
	}

	if *listenAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	go func() {
		for {
			file, err := os.Open("/etc/bird/bird.conf")
			if err != nil {
				log.Fatal(err)
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() { // read first line from buffer
				elems := strings.Split(scanner.Text(), " ")
				if len(elems) != 4 {
					log.Fatalln("unable to parse bcg timestamp header, are you running bcg >= 1.0.63?")
				}
				timestamp, err := strconv.Atoi(elems[3])
				if err != nil {
					log.Fatalln("unable to parse bcg timestamp header, are you running bcg >= 1.0.63?")
				}
				lastUpdate.Set(float64(timestamp))
				break
			}

			_ = file.Close()

			time.Sleep(updateDelay)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting exporter: http://%s/metrics", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
