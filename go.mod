module github.com/kitex-contrib/registry-consul

go 1.16

require (
	github.com/apache/thrift v0.20.0
	github.com/cloudwego-contrib/cwgo-pkg/registry/consul v0.0.0-00010101000000-000000000000
	github.com/cloudwego/kitex v0.11.0
	github.com/hashicorp/consul/api v1.26.1
	github.com/stretchr/testify v1.9.0
)

replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

replace github.com/cloudwego-contrib/cwgo-pkg/registry/consul => github.com/smx-Morgan/cwgo-pkg/registry/consul v0.0.0-20241016000926-d56ef7e0f578
