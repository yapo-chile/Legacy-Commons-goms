package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.schibsted.io/Yapo/goms/pkg/infrastructure"
	"github.schibsted.io/Yapo/goms/pkg/interfaces/handlers"
	// CLONE REMOVE START
	"github.schibsted.io/Yapo/goms/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/goms/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/goms/pkg/usecases"
	// CLONE REMOVE END
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

	fmt.Printf("Setting up Prometheus\n")
	prometheus := infrastructure.MakePrometheusExporter(
		conf.PrometheusConf.Port,
		conf.PrometheusConf.Enabled,
	)

	fmt.Printf("Setting up logger\n")
	logger, err := infrastructure.MakeYapoLogger(&conf.LoggerConf,
		prometheus.NewEventsCollector(
			"goms_service_events_total",
			"events tracker counter for goms service",
		),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	shutdownSequence.Push(prometheus)

	logger.Info("Initializing resources")

	// HealthHandler
	var healthHandler handlers.HealthHandler

	// CLONE REMOVE START
	// FibonacciHandler
	fibonacciLogger := loggers.MakeFibonacciLogger(logger)
	fibonacciRepository := repository.NewMapFibonacciRepository()
	fibonacciInteractor := usecases.FibonacciInteractor{
		Logger:     fibonacciLogger,
		Repository: fibonacciRepository,
	}
	fibonacciHandler := handlers.FibonacciHandler{
		Interactor: &fibonacciInteractor,
	}
	// CLONE REMOVE END

	// Setting up router
	maker := infrastructure.RouterMaker{
		Logger:        logger,
		WrapperFuncs:  []infrastructure.WrapperFunc{prometheus.TrackHandlerFunc},
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
					// CLONE REMOVE START
					{
						Name:    "Retrieve the Nth Fibonacci with Clean Architecture",
						Method:  "GET",
						Pattern: "/fibonacci",
						Handler: &fibonacciHandler,
					},
					// CLONE REMOVE END
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
