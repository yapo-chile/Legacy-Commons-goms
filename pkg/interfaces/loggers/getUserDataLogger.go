package loggers

import (
	"github.mpi-internal.com/Yapo/goms/pkg/usecases"
)

type getUserPrometheusDefaultLogger struct {
	logger Logger
}

func (l *getUserPrometheusDefaultLogger) LogBadInput(n string) {
	l.logger.Error("Wrong input type: %d", n)
}

// MakeGetUserPrometheusDefaultLogger sets up a userProfileLogger instrumented via the provided logger
func MakeGetUserPrometheusDefaultLogger(logger Logger) usecases.GetUserPrometheusDefaultLogger {
	return &getUserPrometheusDefaultLogger{
		logger: logger,
	}
}