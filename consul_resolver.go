/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package consul

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// NewConsulResolverWithConfig create a service resolver using consul, with a custom config.
func NewConsulResolverWithConfig(config *api.Config) (discovery.Resolver, error) {
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
	agentServiceList, _, err := c.consulClient.Health().Service(desc, "", true, nil)
	if err != nil {
		return discovery.Result{}, err
	}
	if len(agentServiceList) == 0 {
		return discovery.Result{}, errors.New("no service found")
	}

	for _, i := range agentServiceList {
		svc := i.Service
		if svc == nil || svc.Address == "" {
			continue
		}

		eps = append(eps, discovery.NewInstance(
			defaultNetwork,
			fmt.Sprint(svc.Address, ":", svc.Port),
			svc.Weights.Passing,
			splitTags(svc.Tags),
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

// splitTags Tags characters be separated to map.
func splitTags(tags []string) map[string]string {
	n := len(tags)
	tagMap := make(map[string]string, n)
	if n == 0 {
		return tagMap
	}

	for _, tag := range tags {
		if tag == "" {
			continue
		}
		strArr := strings.SplitN(tag, kvJoinChar, 2)
		if len(strArr) == 2 {
			key := strArr[0]
			tagMap[key] = strArr[1]
		}
	}

	return tagMap
}
