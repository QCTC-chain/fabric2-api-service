/*
Copyright Suzhou Tongji Fintech Research Institute 2017 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gmcredentials

// import gmtls "github.com/Hyperledger-TWGC/tjfoc-gm/tls"
import gmtls "gitee.com/china_uni/tjfoc-gm/tls"

// cloneTLSConfig returns a shallow clone of the exported
// fields of cfg, ignoring the unexported sync.Once, which
// contains a mutex and must not be copied.
//
// If cfg is nil, a new zero tls.Config is returned.
func cloneTLSConfig(cfg *gmtls.Config) *gmtls.Config {
	if cfg == nil {
		return &gmtls.Config{}
	}

	return cfg.Clone()
}
