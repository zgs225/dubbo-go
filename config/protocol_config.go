/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"strings"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common/constant"
)

// ProtocolConfig is protocol configuration
type ProtocolConfig struct {
	Name string `default:"dubbo" validate:"required" yaml:"name"  json:"name,omitempty" property:"name"`
	Ip   string `default:"127.0.0.1" yaml:"ip"  json:"ip,omitempty" property:"ip"`
	Port int    `default:"0" yaml:"port" json:"port,omitempty" property:"port"`
}

// Prefix dubbo.protocols
func (ProtocolConfig) Prefix() string {
	return constant.ProtocolConfigPrefix
}

// getProtocolsConfig get protocols config default protocol
func getProtocolsConfig(protocols map[string]*ProtocolConfig) map[string]*ProtocolConfig {
	if protocols == nil || len(protocols) <= 0 {
		conf := new(ProtocolConfig)
		protocols = make(map[string]*ProtocolConfig, 1)
		protocols[constant.DEFAULT_Key] = conf
	}
	return protocols
}

func loadProtocol(protocolsIds string, protocols map[string]*ProtocolConfig) []*ProtocolConfig {
	returnProtocols := make([]*ProtocolConfig, 0, len(protocols))
	for _, v := range strings.Split(protocolsIds, ",") {
		for k, protocol := range protocols {
			if v == k {
				returnProtocols = append(returnProtocols, protocol)
			}
		}
	}
	return returnProtocols
}
