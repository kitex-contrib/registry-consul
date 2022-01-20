package consul

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hashicorp/consul/api"
)

const (
	weightMetaKey = "weight"
	defaultWeight = 10
)

// consulHttpResolver is a consul resolver based on HTTP protocol.
type consulHttpResolver struct {
	consulClient *api.Client
}

// ConsulResolverConfig is used to configure the creation of a client
type ConsulResolverConfig struct {
	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	ACLToken string

	// generate a TLSClientConfig that's useful for talking to Consul using TLS.
	TLSConfig *api.TLSConfig

	// HTTPAuth is used to authenticate http client with HTTP Basic Authentication
	HTTPAuth *api.HttpBasicAuth
}

// NewConsulResolver creates a consul based resolver.
func NewConsulResolver(endpoint string, extraConfig *ConsulResolverConfig) (discovery.Resolver, error) {
	// Make client config
	conf := api.DefaultConfig()

	conf.Address = endpoint

	if extraConfig != nil {
		conf.Datacenter = extraConfig.Datacenter
		conf.Token = extraConfig.ACLToken

		if extraConfig.TLSConfig != nil {
			conf.TLSConfig = *extraConfig.TLSConfig
		}

		if extraConfig.HTTPAuth != nil {
			conf.HttpAuth = extraConfig.HTTPAuth
		}
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return &consulHttpResolver{
		consulClient: client,
	}, nil
}

// Name implements the Resolver interface.
func (r *consulHttpResolver) Name() string {
	return "consul"
}

// Target implements the Resolver interface.
func (r *consulHttpResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve implements the Resolver interface.
func (r *consulHttpResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	agentServices, err := r.consulClient.Agent().Services()
	if err != nil {
		return discovery.Result{}, err
	}

	var eps []discovery.Instance
	for _, service := range agentServices {
		address := fmt.Sprintf("%s:%d", service.Address, service.Port)
		weight := queryWeight(service)
		eps = append(eps, discovery.NewInstance("tcp", address, weight, service.Meta))
	}

	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

// Diff implements the Resolver interface.
func (e *consulHttpResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

func queryWeight(service *api.AgentService) int {
	var weight int
	var err error

	if weightValue, ok := service.Meta[weightMetaKey]; ok {
		if weight, err = strconv.Atoi(weightValue); err != nil {
			weight = defaultWeight
		}
	} else {
		weight = defaultWeight
	}
	return weight
}
