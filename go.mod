module github.com/qctc/fabric2-api-server

go 1.14

require (
	github.com/gorilla/mux v1.8.1
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitee.com/china_uni/tjfoc-gm v1.2.1 => ./third_party/tjfoc-gm
	github.com/hyperledger/fabric-sdk-go => ./third_party/fabric-sdk-go
)
