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
// See the License for the specific

package test

import (
	"context"
	"net"
	"testing"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/stretchr/testify/assert"
)

func TestConsulDiscovery(t *testing.T) {
	// register
	r, err := consul.NewConsulRegister("127.0.0.1:8500")
	assert.Nil(t, err)
	tags := map[string]string{"group": "blue", "idc": "hd1"}
	addr, _ := net.ResolveTCPAddr("tcp", ":9999")
	info := &registry.Info{ServiceName: "product", Weight: 100, PayloadCodec: "thrift", Tags: tags, Addr: addr}
	err = r.Register(info)
	assert.Nil(t, err)

	// resolve
	res, err := consul.NewConsulResolver("127.0.0.1:8500")
	assert.Nil(t, err)
	target := res.Target(context.Background(), rpcinfo.NewEndpointInfo("product", "", nil, nil))
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)

	// compare data
	if len(result.Instances) == 0 {
		t.Errorf("instance num mismatch, expect: %d, in fact: %d", 1, 0)
	} else if len(result.Instances) == 1 {
		instance := result.Instances[0]
		host, port, err := net.SplitHostPort(instance.Address().String())
		assert.Nil(t, err)
		local, _ := consul.GetLocalIPv4Address()
		if host != local {
			t.Errorf("instance host is mismatch, expect: %s, in fact: %s", local, host)
		}
		if port != "9999" {
			t.Errorf("instance port is mismatch, expect: %s, in fact: %s", "9999", port)
		}
		if info.Weight != instance.Weight() {
			t.Errorf("instance weight is mismatch, expect: %d, in fact: %d", info.Weight, instance.Weight())
		}
		for k, v := range info.Tags {
			if v1, exist := instance.Tag(k); !exist || v != v1 {
				t.Errorf("instance tags is mismatch, expect k:v %s:%s, in fact k:v %s:%s", k, v, k, v1)
			}
		}
	}

	// deregister
	err = r.Deregister(info)
	assert.Nil(t, err)

	// resolve again
	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	if len(result.Instances) != 0 {
		t.Errorf("instance num mismatch, expect: %d, in fact: %d", 0, len(result.Instances))
	}
}
