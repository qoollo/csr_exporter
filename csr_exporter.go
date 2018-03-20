package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"github.com/prometheus/client_golang/prometheus"
)

var log = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)


var backendLeveled logging.LeveledBackend
func init() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendLeveled = logging.AddModuleLevel(logging.NewBackendFormatter(backend, format))
	logging.SetBackend(backendLeveled)

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
}

type metric struct {
	Name  string
	Help  string
	Cmd   string
	gauge prometheus.Gauge
	hasValue bool
}

type configToml struct {
	UpdatePeriodSec int `toml:"update_period_sec"`
	Port            int `toml:"port"`
	Metrics         []metric
}

func updateMetricCmd(m *metric) {
	out, err := exec.Command("sh", "-c", m.Cmd).Output()
	if err != nil {
		log.Error("Error in executing - %s: %s", m.Cmd, err)
		if m.hasValue {
			prometheus.Unregister(m.gauge)
			m.hasValue = false
		}
	}	else {
		cleanOut := strings.TrimRight(string(out), "\n")
		log.Debug("%s = %s", m.Name, cleanOut)

		v, err := strconv.ParseFloat(cleanOut, 64)
		if err != nil {
			log.Error("Error in executing - %s: %s", m.Cmd, err)
			prometheus.Unregister(m.gauge)
			m.hasValue = false
		} else {
			if !m.hasValue {
				prometheus.MustRegister(m.gauge)
				m.hasValue = true
			}
			m.gauge.Set(v)
		}
	}
}

func updateMetrics(ms []metric, updatePeriod int) {
	for {
		for i, _ := range ms {
			updateMetricCmd(&ms[i])
		}

		time.Sleep(time.Duration(updatePeriod) * time.Second)
	}
}

func main() {
	var configPath string
	var verbose bool

	flag.StringVar(&configPath, "config", "", "Path to config. (Required)")
	flag.BoolVar(&verbose, "verbose", false, "Enable debug output.")
	flag.Parse()

	if configPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if verbose {
		backendLeveled.SetLevel(logging.DEBUG, "")
	} else {
		backendLeveled.SetLevel(logging.ERROR, "")
	}

	var config configToml
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Critical("%s", err)
		os.Exit(1)
	}

	ms := make([]metric, 0, len(config.Metrics))

	for _, v := range config.Metrics {
		v.gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: v.Name,
			Help: v.Help,
		})

		v.hasValue = false;
		ms = append(ms, v)
	}

	go updateMetrics(ms, config.UpdatePeriodSec)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", prometheus.UninstrumentedHandler())
	log.Debug("Running on port: %d Update period secs: %d", config.Port, config.UpdatePeriodSec)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
