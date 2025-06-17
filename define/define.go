package define

var GlobalConfig *Config

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Fabric map[string]FabricNetwork `yaml:"fabric"`
}

type FabricNetwork struct {
	ConfigFilePath string `yaml:"configFilePath"` // 网络sdk配置目录（如 gRPC URL）
}

type RawConfig struct {
	Version       string                 `yaml:"version"`
	Client        ClientSection          `yaml:"client"`
	Channels      map[string]Channel     `yaml:"channels"`
	Organizations map[string]interface{} `yaml:"organizations"`
	Orderers      map[string]PeerNode    `yaml:"orderers"`
	Peers         map[string]PeerNode    `yaml:"peers"`
}

type ClientSection struct {
	Organization string `yaml:"organization"`
	Logging      struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
	TLSCerts struct {
		SystemCertPool bool `yaml:"systemCertPool"`
		Client         struct {
			Key  KeyPath `yaml:"key"`
			Cert KeyPath `yaml:"cert"`
		} `yaml:"client"`
	} `yaml:"tlsCerts"`
}

type Channel struct {
	Peers map[string]PeerConfig `yaml:"peers"`
}

type PeerConfig struct {
	EndorsingPeer  bool `yaml:"endorsingPeer"`
	ChaincodeQuery bool `yaml:"chaincodeQuery"`
	LedgerQuery    bool `yaml:"ledgerQuery"`
	EventSource    bool `yaml:"eventSource"`
}

type Organization struct {
	Mspid                  string          `yaml:"mspid"`
	Peers                  []string        `yaml:"peers"`
	CertificateAuthorities []string        `yaml:"certificateAuthorities"`
	Users                  map[string]User `yaml:"users"`
}

type OrdererOrg struct {
	MspID string          `yaml:"mspID"`
	Users map[string]User `yaml:"users"`
}

type PeerNode struct {
	URL         string                 `yaml:"url"`
	GRPCOptions map[string]interface{} `yaml:"grpcOptions"`
	TLSCACerts  KeyPath                `yaml:"tlsCACerts"`
}

type KeyPath struct {
	Path string `yaml:"path"`
}

type User struct {
	Cert KeyPath `yaml:"cert"`
	Key  KeyPath `yaml:"key"`
}

type UserConfigRequest struct {
	MspId    string   `json:"mspId"`    // 组织的 MSP ID，例如 "Org1MSP"
	UserName string   `json:"userName"` // 用户名，例如 "Admin" 或 "User1"
	PathId   string   `json:"pathId"`   // 用户证书文件路径ID
	Peers    []string `json:"peers"`    // 可选：该用户所属组织的 peers 列表
}

// ContractInvokeRequest 合约调用请求参数
type ContractInvokeRequest struct {
	UserConfig  UserConfigRequest `json:"userConfig"`
	ChainName   string            `json:"chainName"`
	ChannelID   string            `json:"channelId"`
	ChaincodeID string            `json:"chaincodeId"`
	Method      string            `json:"method"`
	Args        [][]byte          `json:"args"`
}

// ContractQueryRequest 合约查询请求参数
type ContractQueryRequest struct {
	UserConfig  UserConfigRequest `json:"userConfig"`
	ChainName   string            `json:"chainName"`
	ChannelID   string            `json:"channelId"`
	ChaincodeID string            `json:"chaincodeId"`
	Method      string            `json:"method"`
	Args        []string          `json:"args"` // 假设是字符串数组，后续可转为字节
}

// ContractEventSubscribeRequest 合约事件订阅请求参数
type ContractEventSubscribeRequest struct {
	UserConfig  UserConfigRequest `json:"userConfig"`
	ChainName   string            `json:"chainName"`
	ChannelID   string            `json:"channelId"`
	ChaincodeID string            `json:"chaincodeId"`
}

type ContractEventUnSubscribeRequest struct {
	UserConfig  UserConfigRequest `json:"userConfig"`
	ChainName   string            `json:"chainName"`
	ChannelID   string            `json:"channelId"`
	ChaincodeID string            `json:"chaincodeId"`
	RegId       string            `json:"regId"`
}
