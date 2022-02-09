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
	registration *api.AgentServiceRegistration
	check        *api.AgentServiceCheck
}

func NewConsulRegister(address string) (*consulRegistry, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulRegistry{consulClient: client, check: defaultCheck()}, nil
}

func defaultCheck() *api.AgentServiceCheck {
	check := new(api.AgentServiceCheck)
	check.Timeout = "5s"
	check.Interval = "5s"
	check.DeregisterCriticalServiceAfter = "30s"

	return check
}

func NewConsulRegisterWithConfig(config *api.Config) (*consulRegistry, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulRegistry{consulClient: client}, nil
}

func (c *consulRegistry) CustomizeRegistration(registration *api.AgentServiceRegistration) {
	c.registration = registration
}

func (c *consulRegistry) CustomizeCheck(check *api.AgentServiceCheck) {
	c.check = check
}

func (c *consulRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}

	if c.registration == nil {
		c.registration = &api.AgentServiceRegistration{
			Name: info.ServiceName,
		}
	} else {
		c.registration.Name = info.ServiceName
	}

	c.registration.ID = fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())

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
		c.registration.Address = ipv4
	} else {
		c.registration.Address = host
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	c.registration.Port = p

	if c.check != nil {
		c.check.TCP = fmt.Sprintf("%s:%d", host, p)
		c.registration.Check = c.check
	}

	c.registration.Meta = info.Tags

	if c.registration.Meta == nil {
		c.registration.Meta = make(map[string]string)
	}

	c.registration.Meta["network"] = info.Addr.Network()
	c.registration.Weights = &api.AgentWeights{
		Passing: info.Weight,
		Warning: 1,
	}

	return c.consulClient.Agent().ServiceRegister(c.registration)
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

func (c *consulRegistry) Deregister(info *registry.Info) error {
	return c.consulClient.Agent().ServiceDeregister(fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String()))
}
