package consul

import (
	"context"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hashicorp/consul/api"
	"log"
)

const (
	defaultWeight = 10
)

type consulResolver struct {
	consulClient *api.Client
}

func (c consulResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

func (c consulResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	var eps []discovery.Instance

	services, err := c.consulClient.Agent().Services()
	if err != nil {
		log.Printf("err:%v", err)
		return discovery.Result{}, err
	}

	for _, service := range services {
		weight := service.Weights.Passing
		if weight <= 0 {
			weight = defaultWeight
		}

		eps = append(eps, discovery.NewInstance(service.Meta["network"], service.Address, weight, service.Meta))
	}

	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

func (c consulResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

func (c consulResolver) Name() string {
	return "consul"
}

func (c *consulRegistry) Deregister(info *registry.Info) error {
	return c.consulClient.Agent().ServiceDeregister(info.ServiceName)
}

func NewConsulResolver(address string) (discovery.Resolver, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulResolver{consulClient: client}, nil
}
