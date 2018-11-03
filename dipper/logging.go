package dipper

import (
	"github.com/op/go-logging"
	"os"
)

var backend = logging.NewLogBackend(os.Stderr, "", 0)
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{module}.%{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}%{color:reset}`,
)
var backendFormatter = logging.NewBackendFormatter(backend, format)
var backendLeveled = logging.AddModuleLevel(backendFormatter)

func init() {
	backendLeveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backendLeveled)
}

// GetLogger : getting a logger for the module
func GetLogger(module string) *logging.Logger {
	return logging.MustGetLogger(module)
}