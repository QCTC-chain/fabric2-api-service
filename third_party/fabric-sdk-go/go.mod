// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/fabric-sdk-go

replace (
	gitee.com/china_uni/tjfoc-gm v1.2.1 => ../tjfoc-gm
)

require (
	gitee.com/china_uni/tjfoc-gm v1.2.1
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-kit/kit v0.8.0
	github.com/golang/protobuf v1.5.2
	github.com/hyperledger/fabric-config v0.0.5
	github.com/hyperledger/fabric-lib-go v1.0.0
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/miekg/pkcs11 v1.0.3
	github.com/mitchellh/mapstructure v1.3.2
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/viper v1.1.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	google.golang.org/grpc v1.31.0
	gopkg.in/yaml.v2 v2.4.0
)

go 1.14
