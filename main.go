package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/nasa9084/go-switchbot"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

type options struct {
	Port *uint `short:"p" long:"port" description:"Port number to listen, default: 8080"`
}

const (
	namespace = "switchbot"
	subsystem = "meter"
)

var labels = []string{"deviceName", "deviceType", "hubDeviceId"}

func handleProbe(client *switchbot.Client, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		params := r.URL.Query()
		target := params.Get("target")
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}

		devices, _, err := client.Device().List(ctx)
		if err != nil {
			slog.Error("failed to get list", "err", err)
			http.Error(w, "failed to get list", http.StatusInternalServerError)
			return
		}

		i := slices.IndexFunc(devices, func(d switchbot.Device) bool {
			return d.ID == target
		})
		if i < 0 {
			http.Error(w, "target not found", http.StatusNotFound)
			return
		}
		device := devices[i]

		status, err := client.Device().Status(ctx, device.ID)
		if err != nil {
			slog.Error("failed to get device status", "target", target, "err", err)
			http.Error(w, "failed to get device status", http.StatusInternalServerError)
			return
		}

		battery := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "battery",
			Help:      "the current battery level, 0-100",
		}, labels)
		battery.WithLabelValues(
			device.Name,
			string(device.Type),
			device.Hub,
		).Set(float64(status.Battery))

		humidity := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "humidity",
			Help:      "humidity percentage",
		}, labels)
		humidity.WithLabelValues(
			device.Name,
			string(device.Type),
			device.Hub,
		).Set(float64(status.Humidity))

		temperature := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "temperature",
			Help:      "temperature in celsius",
		}, labels)
		temperature.WithLabelValues(
			device.Name,
			string(device.Type),
			device.Hub,
		).Set(float64(status.Temperature))

		registry := prometheus.NewRegistry()
		registry.MustRegister(battery, humidity, temperature)

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		flagsErr := err.(*flags.Error)
		if flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var port string
	if opts.Port != nil {
		port = fmt.Sprintf(":%d", *opts.Port)
	} else {
		port = ":8080"
	}

	token := os.Getenv("SWITCHBOT_OPENTOKEN")
	if token == "" {
		logger.Error("Please set SWITCHBOT_OPENTOKEN env variable")
		os.Exit(1)
	}

	secret := os.Getenv("SWITCHBOT_SECRETKEY")
	if secret == "" {
		logger.Error("Please set SWITCHBOT_SECRETKEY env variable")
		os.Exit(1)
	}

	client := switchbot.New(token, secret)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", handleProbe(client, logger))

	logger.Info("Starting server on port " + port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		logger.Error("error", err)
	}
}
