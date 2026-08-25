package telemetry

import (
	"log"
	"time"
)

type Logger struct{}

func (
	Logger,
) Startup(
	message string,
) {
	log.Printf("startup time=%s %s", time.Now().UTC().Format(time.RFC3339), message)
}
func (
	Logger,
) Error(
	component string,
	err error,
) {
	log.Printf("error component=%s err=%v", component, err)
}
func (
	Logger,
) Info(
	component,
	message string,
) {
	log.Printf("info component=%s %s", component, message)
}
