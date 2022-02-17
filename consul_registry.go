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
	"net"
	"strconv"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hashicorp/consul/api"
)

type consulRegistry struct {
	consulClient *api.Client
	check        *api.AgentServiceCheck
}

var _ registry.Registry = (*consulRegistry)(nil)

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(address string) (registry.Registry, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulRegistry{consulClient: client, check: defaultCheck()}, nil
}

// NewConsulRegisterWithConfig create a new registry using consul, with a custom config.
func NewConsulRegisterWithConfig(config *api.Config) (*consulRegistry, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulRegistry{consulClient: client}, nil
}

// CustomizeCheck customize the check of the service.
func (c *consulRegistry) CustomizeCheck(check *api.AgentServiceCheck) {
	c.check = check
}

// Register register a service to consul.
func (c *consulRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	svcInfo := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String()),
		Name: info.ServiceName,
		Meta: info.Tags,
		Weights: &api.AgentWeights{
			Passing: info.Weight,
			Warning: info.Weight,
		},
	}
	if svcInfo.Meta == nil {
		svcInfo.Meta = make(map[string]string)
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return errors.New("parse registry info addr error")
	}
	if port == "" {
		return errors.New("registry info addr missing port")
	}
	if host == "" || host == "::" {
		ipv4, err := getLocalIPv4Address()
		if err != nil {
			return fmt.Errorf("get local ipv4 error, cause %w", err)
		}
		svcInfo.Address = ipv4
	} else {
		svcInfo.Address = host
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	svcInfo.Port = p
	if c.check != nil {
		c.check.TCP = fmt.Sprintf("%s:%d", host, p)
		svcInfo.Check = c.check
	}
	return c.consulClient.Agent().ServiceRegister(svcInfo)
}

// Deregister deregister a service from consul.
func (c *consulRegistry) Deregister(info *registry.Info) error {
	return c.consulClient.Agent().ServiceDeregister(fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String()))
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
	check.DeregisterCriticalServiceAfter = "30s"

	return check
}
