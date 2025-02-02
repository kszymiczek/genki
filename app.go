package genki

import (
	"context"
	"fmt"
	"google.golang.org/grpc/health/grpc_health_v1"
	"sync"

	"github.com/marcoEgger/genki/broker"
	"github.com/marcoEgger/genki/cli"
	"github.com/marcoEgger/genki/logger"
	"github.com/marcoEgger/genki/server"
	"github.com/marcoEgger/genki/server/http"
	genki "github.com/marcoEgger/genki/service"
)

type application struct {
	servers     []server.Server
	httpServers []http.Server
	broker      broker.Broker
	opts        Options
	stopChan    <-chan struct{}
	appContext  context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	flags       *cli.FlagSet
}

func newService(opts ...Option) *application {
	options := newOptions(opts...)

	svc := &application{
		opts:     options,
		stopChan: genki.NewSignalHandler(),
		wg:       sync.WaitGroup{},
		flags:    cli.NewFlagSet(options.Name),
	}

	svc.appContext, svc.cancel = context.WithCancel(context.Background())

	return svc
}

// Name gives the whole thing a name. Good things have names :)
func (svc *application) Name() string {
	return svc.opts.Name
}

// Run will start everything and wait for an os signal to stop.
// - If the HTTP debug server is enabled, it is added to the server list
// - If a broker is configured, Declare() and Consume() are called
// - Every server in the serverlist is started
// - Wait for signal...
func (svc *application) Run(healthServer grpc_health_v1.HealthServer) error {
	defer svc.cancel()

	// add the debug HTTP server if enabled
	if svc.opts.HttpDebugServerEnabled {
		svc.AddHttpServer(http.NewDebugServer(
			svc.opts.HttpDebugServerPort,
		))
	}

	if svc.broker != nil {
		if err := svc.broker.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize broker: %s", err)
		}
		logger.Infof("initialized broker")
		if svc.broker.HasConsumer() {
			svc.wg.Add(1)
			go svc.broker.Consume(&svc.wg)
		}
	}

	// start all registered servers in a goroutine
	for _, srv := range svc.servers {
		svc.wg.Add(1)
		go srv.ListenAndServe(svc.appContext, &svc.wg, healthServer)
	}
	// start all registered http servers in a goroutine
	for _, srv := range svc.httpServers {
		svc.wg.Add(1)
		go srv.ListenAndServe(svc.appContext, &svc.wg)
	}

	// wait for signal handler to fire and shutdown
	<-svc.stopChan
	logger.Info("received OS signal: application is shutting down")
	if svc.broker != nil {
		svc.broker.Disconnect()
	}
	svc.cancel()
	svc.wg.Wait()

	return nil
}

// AddServer registers a new server with the application
func (svc *application) AddServer(srv server.Server) {
	svc.servers = append(svc.servers, srv)
}

// AddHttpServer registers a new http server with the application
func (svc *application) AddHttpServer(srv http.Server) {
	svc.httpServers = append(svc.httpServers, srv)
}

// Add a broker to the application. The broker is invoked in Run().
func (svc *application) RegisterBroker(broker broker.Broker) {
	svc.broker = broker
}

// Opts returns the internal options
func (svc *application) Opts() Options {
	return svc.opts
}
