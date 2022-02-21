# registry-consul (This is a community driven project)

## Docs

### Server

#### Basic Usage
```
import (
    ...
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    consul "github.com/kitex-contrib/registry-consul"
    consulapi "github.com/hashicorp/consul/api"
)

func main() {
    
    r, err := consul.NewConsulRegister("127.0.0.1:8500")
    if err != nil {
        log.Fatal(err)
    }
    
    server := hello.NewServer(new(HelloImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
        ServiceName: "greet.server",
    }))
    err = server.Run()
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Customize Service Check

registry has a default config for service check like

```
check.Timeout = "5s"
check.Interval = "5s"
check.DeregisterCriticalServiceAfter = "1m"
```

you can also use `WithCheck` to modify your config

```
import (
	...
	consul "github.com/kitex-contrib/registry-consul"
	consulapi "github.com/hashicorp/consul/api"
)
func main() {
    ...
	r, err := consul.NewConsulRegister("127.0.0.1:8500", consul.WithCheck(&consulapi.AgentServiceCheck{
            Interval:                       "7s",
            Timeout:                        "5s",
            DeregisterCriticalServiceAfter: "1m",
	}))
}
```

### Client

```
import (
    ...
    "github.com/cloudwego/kitex/client"
    consul "github.com/kitex-contrib/registry-consul"
    ...
)

func main() {
    ...
    r, err := consul.NewConsulResolver("127.0.0.1:8500")
    if err != nil {
        log.Fatal(err)
    }
    client, err := echo.NewClient("greet.server", client.WithResolver(r))
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Example

See Server and Client in example.

## Compatibility

Compatible with consul.

maintained by: [Hanson](https://github.com/hanson)