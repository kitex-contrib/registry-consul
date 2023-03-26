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

package main

import (
	"context"
	"log"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/kitex-contrib/registry-consul/example/hello/kitex_gen/api"
	"github.com/kitex-contrib/registry-consul/example/hello/kitex_gen/api/hello"
)

func main() {
	consulConfig := consulapi.Config{
		Address: "127.0.0.1:8500",
		Token:   "TEST-MY-TOKEN",
	}
	r, err := consul.NewConsulResolverWithConfig(&consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	c := hello.MustNewClient("hello", client.WithResolver(r), client.WithRPCTimeout(time.Second*3))
	ctx := context.Background()
	for {
		resp, err := c.Echo(ctx, &api.Request{Message: "Hello"})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(resp)
		time.Sleep(time.Second)
	}
}
