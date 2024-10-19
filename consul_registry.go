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
	consulregistry "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/options"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hashicorp/consul/api"
)

const kvJoinChar = ":"

// Option is consul option.
type Option = options.Option

// WithCheck is consul registry option to set AgentServiceCheck.
func WithCheck(check *api.AgentServiceCheck) Option {
	return options.WithCheck(check)
}

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(address string, opts ...Option) (registry.Registry, error) {
	return consulregistry.NewConsulRegister(address, opts...)
}

// NewConsulRegisterWithConfig create a new registry using consul, with a custom config.
func NewConsulRegisterWithConfig(config *api.Config, opts ...Option) (registry.Registry, error) {
	return consulregistry.NewConsulRegisterWithConfig(config, opts...)
}
