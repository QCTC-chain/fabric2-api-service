module github.com/qctc/fabric2-api-server

go 1.14

require (
	github.com/gorilla/mux v1.8.1
	github.com/hyperledger/fabric-sdk-go v1.0.0
)

replace (
	gitee.com/china_uni/tjfoc-gm v1.2.1 => ./third_party/tjfoc-gm
	github.com/hyperledger/fabric-sdk-go => ./third_party/fabric-sdk-go
)
