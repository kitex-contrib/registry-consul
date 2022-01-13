// Copyright 2021 CloudWeGo authors.
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

	if host, port, err := net.SplitHostPort(info.Addr.String()); err == nil {
		if port == "" {
			return fmt.Errorf("registry info addr missing port")
		}
		if host == "" || host == "::" {
			ipv4, err := GetLocalIPv4Address()
			if err != nil {
				return fmt.Errorf("get local ipv4 error, cause %w", err)
			}
			c.registration.Address = ipv4
		} else {
			c.registration.Address = host
		}
		port, _ := strconv.Atoi(port)
		c.registration.Port = port
	} else {
		return fmt.Errorf("parse registry info addr error")
	}

	if c.check != nil {
		c.check.TCP = info.Addr.String()
		c.registration.Check = c.check
	}

	c.registration.Meta = info.Tags

	if c.registration.Meta == nil {
		c.registration.Meta = make(map[string]string)
	}

	c.registration.Meta["network"] = info.Addr.Network()
	if info.Weight <= 0 {
		info.Weight = defaultWeight
	}
	c.registration.Weights = &api.AgentWeights{
		Passing: info.Weight,
		Warning: 1,
	}

	if err := c.consulClient.Agent().ServiceRegister(c.registration); err != nil {
		return err
	}

	return nil
}

func validateRegistryInfo(info *registry.Info) error {
	if info.ServiceName == "" {
		return fmt.Errorf("missing service name in Register")
	}
	if info.Addr == nil {
		return fmt.Errorf("missing addr in Register")
	}
	return nil
}