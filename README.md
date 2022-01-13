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
        ServiceName: "echo",
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
check.DeregisterCriticalServiceAfter = "30s"
```

you can also use `CustomizeCheck` to modify your config
```
func main() {
    ...
    r.CustomizeCheck(&consulapi.AgentServiceCheck{
        Timeout:                        "1s",
        Interval:                       "3s",
        DeregisterCriticalServiceAfter: "10s",
    })
}
```

#### Customize Registration
```
func main() {
    ...
    r.CustomizeRegistration(&consulapi.AgentServiceRegistration{
        Tags: []string{"dev"},
    })
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
    client, err := echo.NewClient("echo", client.WithResolver(r))
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Example

See Server and Client.