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
	"errors"
	"fmt"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hashicorp/consul/api"
)

type consulRegistry struct {
	consulClient *api.Client
	opts         options
}

var _ registry.Registry = (*consulRegistry)(nil)

type options struct {
	check *api.AgentServiceCheck
}

// Option is consul option.
type Option func(o *options)

// WithCheck is consul registry option to set AgentServiceCheck.
func WithCheck(check *api.AgentServiceCheck) Option {
	return func(o *options) { o.check = check }
}

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(address string, opts ...Option) (registry.Registry, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	op := options{
		check: defaultCheck(),
	}

	for _, option := range opts {
		option(&op)
	}

	return &consulRegistry{consulClient: client, opts: op}, nil
}

// NewConsulRegisterWithConfig create a new registry using consul, with a custom config.
func NewConsulRegisterWithConfig(config *api.Config) (*consulRegistry, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulRegistry{consulClient: client}, nil
}

// Register register a service to consul.
func (c *consulRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}

	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return err
	}
	svcID, err := getServiceId(info)
	if err != nil {
		return err
	}
	svcInfo := &api.AgentServiceRegistration{
		ID:      svcID,
		Address: host,
		Port:    port,
		Name:    info.ServiceName,
		Meta:    info.Tags,
		Weights: &api.AgentWeights{
			Passing: info.Weight,
			Warning: info.Weight,
		},
		Check: c.opts.check,
	}

	if c.opts.check != nil {
		c.opts.check.TCP = fmt.Sprintf("%s:%d", host, port)
		svcInfo.Check = c.opts.check
	}

	return c.consulClient.Agent().ServiceRegister(svcInfo)
}

// Deregister deregister a service from consul.
func (c *consulRegistry) Deregister(info *registry.Info) error {
	svcID, err := getServiceId(info)
	if err != nil {
		return err
	}
	return c.consulClient.Agent().ServiceDeregister(svcID)
}

func validateRegistryInfo(info *registry.Info) error {
	if info.ServiceName == "" {
		return errors.New("missing service name in consul register")
	}
	if info.Addr == nil {
		return errors.New("missing addr in consul register")
	}
	return nil
}

func defaultCheck() *api.AgentServiceCheck {
	check := new(api.AgentServiceCheck)
	check.Timeout = "5s"
	check.Interval = "5s"
	check.DeregisterCriticalServiceAfter = "1m"

	return check
}
