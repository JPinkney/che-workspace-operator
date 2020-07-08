package prestop

import (
	"os"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt}
var log = logf.Log.WithName("prestop")

func SetupSignalHandler() (stopCh <-chan struct{}) {
	log.Info("Am i being called?")
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		log.Info("Removing resources")
		RemoveExistingCustomResources()
//		close(stop)
//		<-c
//		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}