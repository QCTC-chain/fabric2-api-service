package define

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"sync"
)

var (
	GlobalConfig   *Config
	GlobalProducer rocketmq.Producer

	EventSubscriptions = make(map[string]fab.Registration) // key: uuid
	SubscriptionMutex  = &sync.RWMutex{}
)

type MQConfig struct {
	Type     string `yaml:"type"`     // 消息队列类型（如 rocketmq）
	Host     string `yaml:"host"`     // 主机地址
	Port     int    `yaml:"port"`     // 端口
	UserName string `yaml:"userName"` // 用户名
	Password string `yaml:"password"` // 密码
	Topic    string `yaml:"topic"`    // 主题
	Group    string `yaml:"group"`    // 消费组
}

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	ChainType string `yaml:"chainType"`

	MQ MQConfig `yaml:"mq"` // 添加 mq 的配置
}

type SdkConfigRequest struct {
	SdkConfig string `json:"sdkConfig"`
}

// ContractInvokeRequest 合约调用请求参数
type ContractInvokeRequest struct {
	SdkConfig     string   `json:"sdkConfig"`
	ChaincodeName string   `json:"chaincodeName"`
	Method        string   `json:"method"`
	Args          []string `json:"args"`
}

// ContractQueryRequest 合约查询请求参数
type ContractQueryRequest struct {
	SdkConfig     string   `json:"sdkConfig"`
	ChaincodeName string   `json:"ChaincodeName"`
	Method        string   `json:"method"`
	Args          []string `json:"args"` // 假设是字符串数组，后续可转为字节
}

// ContractEventSubscribeRequest 合约事件订阅请求参数
type ContractEventSubscribeRequest struct {
	SdkConfig     string `json:"sdkConfig"`
	ChaincodeName string `json:"chaincodeName"`
	EventName     string `json:"eventName"`
}

type ContractEventUnSubscribeRequest struct {
	SdkConfig     string `json:"sdkConfig"`
	ChaincodeName string `json:"chaincodeName"`
	EventName     string `json:"eventName"`
}

type ContractListRequest struct {
	SdkConfig     string `json:"sdkConfig"`
	ChaincodeName string `json:"chaincodeName"`
}

type GetBlockRequest struct {
	SdkConfig   string `json:"sdkConfig"`
	BlockNumber string `json:"blockNumber"`
	OnlyHeader  bool   `json:"onlyHeader"`
}

type GetTxRequest struct {
	SdkConfig   string `json:"sdkConfig"`
	TxId        string `json:"txId"`
	BlockNumber uint64 `json:"blockNumber"`
	IsVerified  bool   `json:"isVerified"`
}
