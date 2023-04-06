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
	"log"
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
	consulAddr  = "127.0.0.1:8500"
	consulToken = "TEST-MY-TOKEN1"
)

var (
	consulClient *consulapi.Client
	cRegistry    registry.Registry
	cResolver    discovery.Resolver
	localIpAddr  string
)

func init() {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	c, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
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

	localIpAddr, _ = getLocalIPv4Address()
}

// TestNewConsulRegister tests the NewConsulRegister function.
func TestNewConsulRegister(t *testing.T) {
	_, err := NewConsulRegister(consulAddr)
	assert.NoError(t, err)
}

// TestNewConsulRegisterWithConfig tests the NewConsulRegisterWithConfig function.
func TestNewConsulRegisterWithConfig(t *testing.T) {
	_, err := NewConsulRegisterWithConfig(&consulapi.Config{
		Address:   consulAddr,
		WaitTime:  5 * time.Second,
		Namespace: "TEST-NS",
	}, WithCheck(&consulapi.AgentServiceCheck{
		Interval:                       "7s",
		Timeout:                        "5s",
		DeregisterCriticalServiceAfter: "15s",
	}))
	assert.NoError(t, err)
}

// TestNewConsulResolver tests the NewConsulResolver function.
func TestNewConsulResolver(t *testing.T) {
	_, err := NewConsulResolver(consulAddr)
	assert.NoError(t, err)
}

// TestNewConsulResolverWithConfig tests the NewConsulResolverWithConfig function.
func TestNewConsulResolverWithConfig(t *testing.T) {
	consulResolverWithConfig, err := NewConsulResolverWithConfig(&consulapi.Config{
		Address: consulAddr,
		Token:   consulToken,
	})
	assert.NoError(t, err)
	assert.NotNil(t, consulResolverWithConfig)
}

// TestNewConsulResolver tests unit test preparatory work.
func TestConsulPrepared(t *testing.T) {
	assert.NotNil(t, consulClient)
	assert.NotNil(t, cRegistry)
	assert.NotNil(t, cResolver)
}

// TestRegister tests the Register function.
func TestRegister(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcPort   = 8081
		testSvcWeight = 777
		tagMap        = map[string]string{
			"k1": "vv1",
			"k2": "vv2",
			"k3": "vv3",
		}
		tagList = []string{"k1:vv1", "k2:vv2", "k3:vv3"}
	)

	// listen on the port, and wait for the health check to connect
	addr := fmt.Sprintf("%s:%d", localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr)
		t.Fail()
	}
	defer lis.Close()

	testSvcAddr, _ := net.ResolveTCPAddr("tcp", addr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
		Tags:        tagMap,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	// wait for health check passing
	time.Sleep(time.Second * 6)

	list, _, err := consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(list)) {
		ss := list[0]
		gotSvc := ss.Service
		assert.Equal(t, testSvcName, gotSvc.Service)
		assert.Equal(t, testSvcAddr.String(), fmt.Sprintf("%s:%d", gotSvc.Address, gotSvc.Port))
		assert.Equal(t, testSvcWeight, gotSvc.Weights.Passing)
		assert.Equal(t, tagList, gotSvc.Tags)
	}
}

// TestConsulDiscovery tests the ConsulDiscovery function.
func TestConsulDiscovery(t *testing.T) {
	var (
		testSvcName = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"

		testSvcPort   = 8082
		testSvcWeight = 777

		ctx = context.Background()

		tagMap = map[string]string{
			"k1": "v1",
			"k2": "v22",
			"k3": "v333",
		}
	)
	// listen on the port, and wait for the health check to connect
	addr := fmt.Sprintf("%s:%d", localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr)
		t.Fail()
	}
	defer lis.Close()

	testSvcAddr, _ := net.ResolveTCPAddr("tcp", addr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
		Tags:        tagMap,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	// wait for health check passing
	time.Sleep(time.Second * 6)

	// resolve
	result, err := cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(result.Instances)) {
		instance := result.Instances[0]
		assert.Equal(t, testSvcWeight, instance.Weight())
		assert.Equal(t, testSvcAddr.String(), instance.Address().String())
		v1, ok := instance.Tag("k1")
		if assert.Equal(t, ok, true) {
			assert.Equal(t, "v1", v1)
		}
		v2, ok := instance.Tag("k2")
		if assert.Equal(t, ok, true) {
			assert.Equal(t, "v22", v2)
		}
		v3, ok := instance.Tag("k3")
		if assert.Equal(t, ok, true) {
			assert.Equal(t, "v333", v3)
		}
	}
}

// TestDeregister tests the Deregister function.
func TestDeregister(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcPort   = 8083
		testSvcWeight = 777
		ctx           = context.Background()
	)
	// listen on the port, and wait for the health check to connect
	addr := fmt.Sprintf("%s:%d", localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr)
		t.Fail()
	}
	defer lis.Close()

	testSvcAddr, _ := net.ResolveTCPAddr("tcp", addr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	time.Sleep(time.Second * 6)

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

// TestMultiInstancesRegister tests the Register function, register multiple instances, then deregister one of them.
func TestMultiInstancesRegister(t *testing.T) {
	var (
		testSvcName = "svc.local"

		testPort1 = 8091
		testPort2 = 8092
		testPort3 = 8093
	)

	addr1 := fmt.Sprintf("%s:%d", localIpAddr, testPort1)
	lis, err := net.Listen("tcp", addr1)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr1)
		t.Fail()
	}
	defer lis.Close()

	testSvcAddr1, _ := net.ResolveTCPAddr("tcp", addr1)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      11,
		Addr:        testSvcAddr1,
	})
	assert.Nil(t, err)

	addr2 := fmt.Sprintf("%s:%d", localIpAddr, testPort2)
	lis2, err := net.Listen("tcp", addr2)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr2)
		t.Fail()
	}
	defer lis2.Close()

	testSvcAddr2, _ := net.ResolveTCPAddr("tcp", addr2)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        testSvcAddr2,
	})
	assert.Nil(t, err)

	addr3 := fmt.Sprintf("%s:%d", localIpAddr, testPort3)
	lis3, err := net.Listen("tcp", addr3)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr3)
		t.Fail()
		return
	}
	defer lis3.Close()

	testSvcAddr3, _ := net.ResolveTCPAddr("tcp", addr3)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      33,
		Addr:        testSvcAddr3,
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 6)

	svcList, _, err := consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(svcList))

	err = cRegistry.Deregister(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        testSvcAddr2,
	})
	assert.Nil(t, err)
	svcList, _, err = consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(svcList)) {
		for _, entry := range svcList {
			gotSvc := entry.Service
			assert.Equal(t, testSvcName, gotSvc.Service)
			assert.Contains(t, []int{testPort1, testPort3}, gotSvc.Port)
			assert.Equal(t, localIpAddr, gotSvc.Address)
		}
	}
}
