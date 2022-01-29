// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consul

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hashicorp/consul/api"
)

type consulResolver struct {
	consulClient *api.Client
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

func (c *consulResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

func (c *consulResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	var eps []discovery.Instance

	services, err := c.consulClient.Agent().Services()
	if err != nil {
		log.Printf("err:%v", err)
		return discovery.Result{}, err
	}
	if len(services) == 0 {
		return discovery.Result{}, errors.New("no service found")
	}

	for _, service := range services {
		log.Println(service.Address)
		weight := service.Weights.Passing
		eps = append(eps, discovery.NewInstance(service.Meta["network"], fmt.Sprintf("%s:%d", service.Address, service.Port), weight, service.Meta))
	}

	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

func (c *consulResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

func (c *consulResolver) Name() string {
	return "consul"
}

func (c *consulRegistry) Deregister(info *registry.Info) error {
	return c.consulClient.Agent().ServiceDeregister(info.ServiceName)
}
