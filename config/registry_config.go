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
	"net/url"
	"strconv"
	"strings"
)

import (
	"github.com/creasty/defaults"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/logger"
)

// RegistryConfig is the configuration of the registry center
type RegistryConfig struct {
	Protocol string `default:"zookeeper" validate:"required" yaml:"protocol"  json:"protocol,omitempty" property:"protocol"`
	Timeout  string `default:"10s" validate:"required" yaml:"timeout" json:"timeout,omitempty" property:"timeout"` // unit: second
	Group    string `yaml:"group" json:"group,omitempty" property:"group"`
	TTL      string `default:"10m" yaml:"ttl" json:"ttl,omitempty" property:"ttl"` // unit: minute
	// for registry
	Address    string `default:"zookeeper://127.0.0.1:2181" validate:"required" yaml:"address" json:"address,omitempty" property:"address"`
	Username   string `yaml:"username" json:"username,omitempty" property:"username"`
	Password   string `yaml:"password" json:"password,omitempty"  property:"password"`
	Simplified bool   `yaml:"simplified" json:"simplified,omitempty"  property:"simplified"`
	// Always use this registry first if set to true, useful when subscribe to multiple registries
	Preferred bool `yaml:"preferred" json:"preferred,omitempty" property:"preferred"`
	// The region where the registry belongs, usually used to isolate traffics
	Zone string `yaml:"zone" json:"zone,omitempty" property:"zone"`
	// Affects traffic distribution among registries,
	// useful when subscribe to multiple registries Take effect only when no preferred registry is specified.
	Weight int64             `yaml:"weight" json:"weight,omitempty" property:"weight"`
	Params map[string]string `yaml:"params" json:"params,omitempty" property:"params"`
}

func getRegistriesConfig(registries map[string]*RegistryConfig) map[string]*RegistryConfig {
	if registries == nil || len(registries) <= 0 {
		registries = make(map[string]*RegistryConfig, 1)
		reg := new(RegistryConfig)
		registries[constant.DEFAULT_Key] = reg
		return registries
	}
	return registries
}

// UnmarshalYAML unmarshal the RegistryConfig by @unmarshal function
func (c *RegistryConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	type plain RegistryConfig
	return unmarshal((*plain)(c))
}

// Prefix dubbo.registries
func (RegistryConfig) Prefix() string {
	return constant.RegistryConfigPrefix
}

func loadRegistries(targetRegistries string, registries map[string]*RegistryConfig, roleType common.RoleType) []*common.URL {
	var urls []*common.URL
	trSlice := strings.Split(targetRegistries, ",")

	for k, registryConf := range registries {
		target := false

		// if user not config targetRegistries, default load all
		// Notice: in func "func Split(s, sep string) []string" comment:
		// if s does not contain sep and sep is not empty, SplitAfter returns
		// a slice of length 1 whose only element is s. So we have to add the
		// condition when targetRegistries string is not set (it will be "" when not set)
		if len(trSlice) == 0 || (len(trSlice) == 1 && trSlice[0] == "") {
			target = true
		} else {
			// else if user config targetRegistries
			for _, tr := range trSlice {
				if tr == k {
					target = true
					break
				}
			}
		}

		if target {
			addresses := strings.Split(registryConf.Address, ",")
			address := addresses[0]
			address = registryConf.translateRegistryAddress()
			url, err := common.NewURL(constant.REGISTRY_PROTOCOL+"://"+address,
				common.WithParams(registryConf.getUrlMap(roleType)),
				common.WithParamsValue("simplified", strconv.FormatBool(registryConf.Simplified)),
				common.WithUsername(registryConf.Username),
				common.WithPassword(registryConf.Password),
				common.WithLocation(registryConf.Address),
			)

			if err != nil {
				logger.Errorf("The registry id: %s url is invalid, error: %#v", k, err)
				panic(err)
			} else {
				urls = append(urls, url)
			}
		}
	}

	return urls
}

func (c *RegistryConfig) getUrlMap(roleType common.RoleType) url.Values {
	urlMap := url.Values{}
	urlMap.Set(constant.GROUP_KEY, c.Group)
	urlMap.Set(constant.ROLE_KEY, strconv.Itoa(int(roleType)))
	urlMap.Set(constant.REGISTRY_KEY, c.Protocol)
	urlMap.Set(constant.REGISTRY_TIMEOUT_KEY, c.Timeout)
	// multi registry invoker weight label for load balance
	urlMap.Set(constant.REGISTRY_KEY+"."+constant.REGISTRY_LABEL_KEY, strconv.FormatBool(true))
	urlMap.Set(constant.REGISTRY_KEY+"."+constant.PREFERRED_KEY, strconv.FormatBool(c.Preferred))
	urlMap.Set(constant.REGISTRY_KEY+"."+constant.ZONE_KEY, c.Zone)
	urlMap.Set(constant.REGISTRY_KEY+"."+constant.WEIGHT_KEY, strconv.FormatInt(c.Weight, 10))
	urlMap.Set(constant.REGISTRY_TTL_KEY, c.TTL)
	for k, v := range c.Params {
		urlMap.Set(k, v)
	}
	return urlMap
}

//translateRegistryAddress translate registry address
//  eg:address=nacos://127.0.0.1:8848 will return 127.0.0.1:8848 and protocol will set nacos
func (c *RegistryConfig) translateRegistryAddress() string {
	if strings.Contains(c.Address, "://") {
		translatedUrl, err := url.Parse(c.Address)
		if err != nil {
			logger.Errorf("The registry url is invalid, error: %#v", err)
			panic(err)
		}
		c.Protocol = translatedUrl.Scheme
		c.Address = strings.Replace(c.Address, translatedUrl.Scheme+"://", "", -1)
	}
	return c.Address
}
