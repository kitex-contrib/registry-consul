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

package consul

import (
	"strings"

	consuleesolver "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/hashicorp/consul/api"
)

const (
	defaultNetwork = "tcp"
)

// NewConsulResolver create a service resolver using consul.
func NewConsulResolver(address string) (discovery.Resolver, error) {
	return consuleesolver.NewConsulResolver(address)
}

// NewConsulResolverWithConfig create a service resolver using consul, with a custom config.
func NewConsulResolverWithConfig(config *api.Config) (discovery.Resolver, error) {
	return consuleesolver.NewConsulResolverWithConfig(config)
}

// splitTags Tags characters be separated to map.
func splitTags(tags []string) map[string]string {
	n := len(tags)
	tagMap := make(map[string]string, n)
	if n == 0 {
		return tagMap
	}

	for _, tag := range tags {
		if tag == "" {
			continue
		}
		strArr := strings.SplitN(tag, kvJoinChar, 2)
		if len(strArr) == 2 {
			key := strArr[0]
			tagMap[key] = strArr[1]
		}
	}

	return tagMap
}
