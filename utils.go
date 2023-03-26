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
	"errors"
	"fmt"
	"net"

	"github.com/cloudwego/kitex/pkg/registry"
)

func getLocalIPv4Address() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("not found ipv4 address")
}

func parseAddr(addr net.Addr) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", 0, err
	}
	if host == "" || host == "::" {
		host, err = getLocalIPv4Address()
		if err != nil {
			return "", 0, fmt.Errorf("get local ipv4 error, cause %w", err)
		}
	}
	port, err = net.LookupPort(defaultNetwork, portStr)
	if err != nil {
		return "", 0, err
	}
	if port == 0 {
		return "", 0, fmt.Errorf("invalid port %s", portStr)
	}

	return host, port, nil
}

func getServiceId(info *registry.Info) (string, error) {
	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}
