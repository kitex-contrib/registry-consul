package main

import (
	api "github.com/kitex-contrib/registry-consul/example/hello/kitex_gen/api/hello"
	"log"
)

func main() {
	svr := api.NewServer(new(HelloImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
