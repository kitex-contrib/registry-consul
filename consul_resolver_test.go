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
// See the License for the specific

package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/sdk/testutil/retry"
	"github.com/stretchr/testify/require"
)

// makeClient create a consul agent client for CRUD.
func makeClient(t *testing.T) (*api.Client, *testutil.TestServer) {
	return makeClientWithConfig(t)
}

func makeClientWithConfig(t *testing.T) (*api.Client, *testutil.TestServer) {
	// Skip test when -short flag provided; any tests that create a test
	// server will take at least 100ms which is undesirable for -short
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	// Make client config
	conf := api.DefaultConfig()

	// Create server
	var server *testutil.TestServer
	var err error
	retry.RunWith(retry.ThreeTimes(), t, func(r *retry.R) {
		server, err = testutil.NewTestServerConfigT(t, nil)
		if err != nil {
			r.Fatalf("Failed to start server: %v", err.Error())
		}
	})
	if server.Config.Bootstrap {
		server.WaitForLeader(t)
	}

	conf.Address = server.HTTPAddr

	// Create client
	client, err := api.NewClient(conf)
	if err != nil {
		server.Stop()
		t.Fatalf("err: %v", err)
	}

	return client, server
}

func registerDemoServices(t *testing.T, agent *api.Agent) {
	reg1 := &api.AgentServiceRegistration{
		Name:    "foo1",
		Port:    8000,
		Address: "192.168.0.42",
	}
	reg2 := &api.AgentServiceRegistration{
		Name: "foo2",
		Port: 8000,
		TaggedAddresses: map[string]api.ServiceAddress{
			"lan": {
				Address: "192.168.0.43",
				Port:    8000,
			},
			"wan": {
				Address: "198.18.0.1",
				Port:    80,
			},
		},
	}
	if err := agent.ServiceRegister(reg1); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := agent.ServiceRegister(reg2); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func Test_ConsulAgentQueryNormal(t *testing.T) {
	c, s := makeClient(t)
	defer s.Stop()

	agent := c.Agent()

	// register some services
	registerDemoServices(t, agent)

	services, err := agent.Services()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if _, ok := services["foo1"]; !ok {
		t.Fatalf("missing service: %v", services)
	}
	if _, ok := services["foo2"]; !ok {
		t.Fatalf("missing service: %v", services)
	}

	if services["foo1"].Address != "192.168.0.42" {
		t.Fatalf("missing Address field in service foo1: %v", services)
	}
	if services["foo2"].Address != "" {
		t.Fatalf("missing Address field in service foo2: %v", services)
	}
	require.NotNil(t, services["foo2"].TaggedAddresses)
	require.Contains(t, services["foo2"].TaggedAddresses, "lan")
	require.Contains(t, services["foo2"].TaggedAddresses, "wan")
	require.Equal(t, services["foo2"].TaggedAddresses["lan"].Address, "192.168.0.43")
	require.Equal(t, services["foo2"].TaggedAddresses["lan"].Port, 8000)
	require.Equal(t, services["foo2"].TaggedAddresses["wan"].Address, "198.18.0.1")
	require.Equal(t, services["foo2"].TaggedAddresses["wan"].Port, 80)

	if err := agent.ServiceDeregister("foo1"); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := agent.ServiceDeregister("foo2"); err != nil {
		t.Fatalf("err: %v", err)
	}
}
