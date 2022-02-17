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
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

const (
	consulAddr = "127.0.0.1:8500"
)

var (
	consulClient *consulapi.Client
	cRegistry    registry.Registry
	cResolver    discovery.Resolver
)

func init() {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	c, err := consulapi.NewClient(config)
	if err != nil {
		return
	}
	consulClient = c

	r, err := NewConsulRegister(consulAddr)
	if err != nil {
		return
	}
	cRegistry = r

	resolver, err := NewConsulResolver(consulAddr)
	if err != nil {
		return
	}
	cResolver = resolver
}

// TestNewConsulRegister tests the NewConsulRegister function.
func TestNewConsulRegister(t *testing.T) {
	_, err := NewConsulRegister(consulAddr)
	assert.NoError(t, err)
}

// NewConsulRegisterWithConfig tests the NewConsulRegisterWithConfig function.
func TestNewConsulRegisterWithConfig(t *testing.T) {
	_, err := NewConsulRegisterWithConfig(&consulapi.Config{
		Address:   consulAddr,
		WaitTime:  5 * time.Second,
		Namespace: "TEST-NS",
	})
	assert.NoError(t, err)
}

// TestNewConsulResolver tests the NewConsulResolver function.
func TestNewConsulResolver(t *testing.T) {
	_, err := NewConsulResolver(consulAddr)
	assert.NoError(t, err)
}

// TestNewConsulResolver tests unit test preparatory work.
func TestConsulPrepared(t *testing.T) {
	assert.NotNil(t, consulClient)
	assert.NotNil(t, cRegistry)
	assert.NotNil(t, cResolver)
}

// TestRegister tests the Register function.
func TestRegister(t *testing.T) {
	svcList, err := consulClient.Agent().Services()
	assert.Nil(t, err)
	svcNum := len(svcList)

	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcAddr   = "127.0.0.1:8080"
		testSvcWeight = 777
		tagList       = map[string]string{
			"k1": "vv1",
			"k2": "vv2",
			"kv": "vv3",
		}
	)
	addr, _ := net.ResolveTCPAddr("tcp", testSvcAddr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        addr,
		Tags:        tagList,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	svcList, err = consulClient.Agent().Services()
	assert.Nil(t, err)
	assert.Equal(t, svcNum+1, len(svcList))

	list, _, err := consulClient.Catalog().Service(testSvcName, "", nil)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(list)) {
		ss := list[0]
		assert.Equal(t, testSvcName, ss.ServiceName)
		assert.Equal(t, testSvcAddr, fmt.Sprintf("%s:%d", ss.ServiceAddress, ss.ServicePort))
		assert.Equal(t, testSvcWeight, ss.ServiceWeights.Passing)
		assert.Equal(t, tagList, ss.ServiceMeta)
	}
}

// TestConsulDiscovery tests the ConsulDiscovery function.
func TestConsulDiscovery(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcAddr   = "127.0.0.1:8181"
		testSvcWeight = 777
		ctx           = context.Background()
	)
	addr, _ := net.ResolveTCPAddr("tcp", testSvcAddr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        addr,
	}
	err := cRegistry.Register(info)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// resolve
	result, err := cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))

	instance := result.Instances[0]
	assert.Equal(t, testSvcWeight, instance.Weight())
	assert.Equal(t, testSvcAddr, instance.Address().String())
}

func TestDeregister(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcAddr   = "127.0.0.1:8181"
		testSvcWeight = 777
		ctx           = context.Background()
	)
	addr, _ := net.ResolveTCPAddr("tcp", testSvcAddr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        addr,
	}
	err := cRegistry.Register(info)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// resolve
	result, err := cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))

	// deregister
	err = cRegistry.Deregister(info)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// resolve again
	result, err = cResolver.Resolve(ctx, testSvcName)
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("no service found"), err)
}

// TestMultiServicesRegister tests the Register function, register multiple services, then deregister one of them.
func TestMultiServicesRegister(t *testing.T) {
	var (
		testSvcName = "svc.local"

		testIP1   = net.IPv4(127, 0, 0, 1)
		testPort1 = 8811

		testIP2   = net.IPv4(127, 0, 0, 2)
		testPort2 = 8822

		testIP3   = net.IPv4(127, 0, 0, 3)
		testPort3 = 8833
	)

	err := cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      11,
		Addr:        &net.TCPAddr{IP: testIP1, Port: testPort1},
	})
	assert.Nil(t, err)

	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        &net.TCPAddr{IP: testIP2, Port: testPort2},
	})
	assert.Nil(t, err)

	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      33,
		Addr:        &net.TCPAddr{IP: testIP3, Port: testPort3},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second)

	svcList, _, err := consulClient.Catalog().Service(testSvcName, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(svcList))

	err = cRegistry.Deregister(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        &net.TCPAddr{IP: testIP2, Port: testPort2},
	})
	svcList, _, err = consulClient.Catalog().Service(testSvcName, "", nil)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(svcList)) {
		for _, service := range svcList {
			assert.Equal(t, testSvcName, service.ServiceName)
			assert.Contains(t, []int{testPort1, testPort3}, service.ServicePort)
			assert.Contains(t, []string{testIP1.String(), testIP3.String()}, service.ServiceAddress)
		}
	}
}
