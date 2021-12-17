package consul

import (
	"fmt"
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
			Name:    info.ServiceName,
			Address: info.Addr.String(),
		}
	} else {
		c.registration.Name = info.ServiceName
		c.registration.Address = info.Addr.String()
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
