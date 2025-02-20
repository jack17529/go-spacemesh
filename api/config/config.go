// Package config provides configuration for GRPC and HTTP api servers
package config

import (
	"errors"
)

const (
	defaultStartGRPCServer    = false
	defaultGRPCServerPort     = 9091
	defaultNewGRPCServerPort  = 9092
	defaultStartJSONServer    = false
	defaultStartNewJSONServer = false
	defaultJSONServerPort     = 9090
	defaultNewJSONServerPort  = 9093
	defaultStartNodeService   = false
	defaultStartMeshService   = false
)

// Config defines the api config params
type Config struct {
	StartGrpcServer    bool     `mapstructure:"grpc-server"`
	StartGrpcServices  []string `mapstructure:"grpc"`
	GrpcServerPort     int      `mapstructure:"grpc-port"`
	NewGrpcServerPort  int      `mapstructure:"grpc-port-new"`
	StartJSONServer    bool     `mapstructure:"json-server"`
	StartNewJSONServer bool     `mapstructure:"json-server-new"`
	JSONServerPort     int      `mapstructure:"json-port"`
	NewJSONServerPort  int      `mapstructure:"json-port-new"`
	// no direct command line flags for these
	StartNodeService bool
	StartMeshService bool
}

func init() {
	// todo: update default config params based on runtime env here
}

// DefaultConfig defines the default configuration options for api
func DefaultConfig() Config {
	return Config{
		StartGrpcServer:    defaultStartGRPCServer, // note: all bool flags default to false so don't set one of these to true here
		StartGrpcServices:  nil,                    // note: cannot configure an array as a const
		GrpcServerPort:     defaultGRPCServerPort,
		NewGrpcServerPort:  defaultNewGRPCServerPort,
		StartJSONServer:    defaultStartJSONServer,
		StartNewJSONServer: defaultStartNewJSONServer,
		JSONServerPort:     defaultJSONServerPort,
		NewJSONServerPort:  defaultNewJSONServerPort,
		StartNodeService:   defaultStartNodeService,
		StartMeshService:   defaultStartMeshService,
	}
}

// ParseServicesList enables the requested services
func (s *Config) ParseServicesList() error {
	// Make sure all enabled GRPC services are known
	for _, svc := range s.StartGrpcServices {
		switch svc {
		case "mesh":
			s.StartMeshService = true
		case "node":
			s.StartNodeService = true
		default:
			return errors.New("unrecognized GRPC service requested: " + svc)
		}
	}

	// If JSON gateway server is enabled, make sure at least one
	// GRPC service is also enabled
	if s.StartNewJSONServer && !s.StartNodeService {
		return errors.New("must enable at least one GRPC service along with JSON gateway service")
	}

	return nil
}
