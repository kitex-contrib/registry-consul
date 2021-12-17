package main

import (
	"context"
	consul "github.com/hanson/registry-consul"
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
