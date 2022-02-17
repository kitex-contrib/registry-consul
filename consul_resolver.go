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
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hashicorp/consul/api"
)

const (
	defaultNetwork = "tcp"
)

type consulResolver struct {
	consulClient *api.Client
}

var _ discovery.Resolver = (*consulResolver)(nil)

// NewConsulResolver create a service resolver using consul.
func NewConsulResolver(address string) (discovery.Resolver, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulResolver{consulClient: client}, nil
}

// Target return a description for the given target that is suitable for being a key for cache.
func (c *consulResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (c *consulResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	var eps []discovery.Instance
	services, _, err := c.consulClient.Catalog().Service(desc, "", nil)
	if err != nil {
		log.Printf("err:%v", err)
		return discovery.Result{}, err
	}
	if len(services) == 0 {
		return discovery.Result{}, errors.New("no service found")
	}
	for _, svc := range services {
		eps = append(eps, discovery.NewInstance(
			defaultNetwork,
			fmt.Sprint(svc.ServiceAddress, ":", svc.ServicePort),
			svc.ServiceWeights.Passing,
			svc.ServiceMeta,
		))
	}

	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

// Diff computes the difference between two results.
func (c *consulResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name return the name of this resolver.
func (c *consulResolver) Name() string {
	return "consul"
}
