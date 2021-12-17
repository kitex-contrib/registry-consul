package main

import (
	"context"
	consul "github.com/hanson/registry-consul"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
)

func main() {
	r, err := consul.NewConsulResolver("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}
	client := hello.MustNewClient("Hello", client.WithResolver(r))
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		resp, err := client.Echo(ctx, &api.Request{Message: "Hello"})
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(resp)
		time.Sleep(time.Second)
	}
}
