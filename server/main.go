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

package main

import (
	"context"
	consul "github.com/kitex-contrib/registry-consul"
	"log"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	consulapi "github.com/hashicorp/consul/api"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message,
	}
	return
}

func main() {
	r, err := consul.NewConsulRegister("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}

	r.CustomizeCheck(&consulapi.AgentServiceCheck{
		Timeout:                        "1s",
		Interval:                       "3s",
		DeregisterCriticalServiceAfter: "10s",
	})

	r.CustomizeRegistration(&consulapi.AgentServiceRegistration{
		Tags: []string{"dev"},
	})

	//r, err := consul.NewConsulRegisterWithConfig(consulapi.DefaultConfig())
	//if err != nil {
	//	log.Fatal(err)
	//}

	server := hello.NewServer(new(HelloImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "Hello",
	}))
	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
