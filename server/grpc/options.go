package grpc

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/lukasjarosch/genki/config"
)

const DefaultPort = "50051"
const DefaultGracePeriod = 3 * time.Second

type Options struct {
	Port                string
	ShutdownGracePeriod time.Duration
}

func Port(addr string) Option {
	return func(opts *Options) {
		opts.Port = addr
	}
}

func ShutdownGracePeriod(duration time.Duration) Option {
	return func(opts *Options) {
		opts.ShutdownGracePeriod = duration
	}
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Port:                DefaultPort,
		ShutdownGracePeriod: DefaultGracePeriod,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Flags is a convenience function to quickly add the gRPC server options as CLI flags
// Implements the cli.FlagProvider type
func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("grpc-server", pflag.ContinueOnError)

	fs.String(
		config.GrpcPort,
		DefaultPort,
		"the port on which the gRPC server is listening on",
	)
	fs.Duration(
		config.GrpcGracePeriod,
		DefaultGracePeriod,
		"grace period after which the server shutdown is terminated",
	)

	return fs
}
