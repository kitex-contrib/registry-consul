# registry-consul

use Consul as service register and discovery backend

## How to use?

### Client
```go
import (
    ...
    consul "github.com/kitex-contrib/registry-consul"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
    ...
    // "127.0.0.1:8500" is the default Consul agent address.
    // DataCenter/ACL/TLS/HTTP basic auth, etc. can use the ConsulResolverConfig to configure.
    r, err := consul.NewConsulResolver("127.0.0.1:8500", &consul.ConsulResolverConfig{})
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

