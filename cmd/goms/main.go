package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.schibsted.io/Yapo/goms/pkg/infrastructure"
	"github.schibsted.io/Yapo/goms/pkg/interfaces/handlers"
	"github.schibsted.io/Yapo/goms/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/goms/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/goms/pkg/usecases"
)

var shutdownSequence = infrastructure.NewShutdownSequence()

func main() {
	var conf infrastructure.Config
	shutdownSequence.Listen()
	infrastructure.LoadFromEnv(&conf)
	if jconf, err := json.MarshalIndent(conf, "", "    "); err == nil {
		fmt.Printf("Config: \n%s\n", jconf)
	} else {
		fmt.Printf("Config: \n%+v\n", conf)
	}

	fmt.Printf("Setting up logger\n")
	logger, err := infrastructure.MakeYapoLogger(&conf.LoggerConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	logger.Info("Setting up New Relic")
	newrelic := infrastructure.NewRelicHandler{
		Appname: conf.NewRelicConf.Appname,
		Key:     conf.NewRelicConf.Key,
		Enabled: conf.NewRelicConf.Enabled,
		Logger:  logger,
	}
	err = newrelic.Start()
	if err != nil {
		logger.Error("Error loading New Relic: %+v", err)
		os.Exit(2)
	}

	logger.Info("Setting up Prometheus")
	prometheus := infrastructure.MakePrometheusExporter(
		conf.PrometheusConf.Port,
		conf.PrometheusConf.Enabled,
		logger,
	)
	shutdownSequence.Push(prometheus)

	metricsExporter := loggers.MakeCustomMetricsExporter(prometheus)

	logger.Info("Initializing resources")

	// HealthHandler
	var healthHandler handlers.HealthHandler
	// FibonacciHandler
	fibonacciLogger := loggers.MakeFibonacciInteractorLogger(logger)
	fibonacciRepository := repository.NewMapFibonacciRepository()
	fibonacciInteractor := usecases.FibonacciInteractor{
		Logger:     fibonacciLogger,
		Repository: fibonacciRepository,
		Metrics:    metricsExporter,
	}
	fibonacciHandler := handlers.FibonacciHandler{
		Interactor: &fibonacciInteractor,
	}

	// Setting up router
	maker := infrastructure.RouterMaker{
		Logger:        logger,
		WrapperFuncs:  []infrastructure.WrapperFunc{newrelic.TrackHandlerFunc, prometheus.TrackHandlerFunc},
		WithProfiling: conf.ServiceConf.Profiling,
		Routes: infrastructure.Routes{
			{
				// This is the base path, all routes will start with this prefix
				Prefix: "/api/v{version:[1-9][0-9]*}",
				Groups: []infrastructure.Route{
					{
						Name:    "Check service health",
						Method:  "GET",
						Pattern: "/healthcheck",
						Handler: &healthHandler,
					},
					{
						Name:    "Retrieve the Nth Fibonacci with Clean Architecture",
						Method:  "GET",
						Pattern: "/fibonacci",
						Handler: &fibonacciHandler,
					},
				},
			},
		},
	}

	router := maker.NewRouter()

	server := infrastructure.NewHTTPServer(
		fmt.Sprintf("%s:%d", conf.Runtime.Host, conf.Runtime.Port),
		router,
		logger,
	)
	shutdownSequence.Push(server)
	logger.Info("Starting request serving")
	go server.ListenAndServe()
	shutdownSequence.Wait()
	logger.Info("Server exited normally")

}
