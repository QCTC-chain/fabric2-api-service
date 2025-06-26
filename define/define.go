package define

import (
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"sync"
)

var (
	GlobalConfig   *Config
	GlobalProducer golang.Producer

	EventSubscriptions  = make(map[string]fab.Registration) // key: uuid
	SubscriptionMutex   = &sync.RWMutex{}
	SubscriptionContext = sync.Map{}
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
	IsGm      bool   `yaml:"isGM"`
	IsSM3     bool   `yaml:"isSM3"`
}

// ContractInvokeRequest 合约调用请求参数
type ContractInvokeRequest struct {
	SdkConfig     string   `json:"sdkConfig"`
	IsGm          bool     `yaml:"isGM"`
	IsSM3         bool     `yaml:"isSM3"`
	ChaincodeName string   `json:"chaincodeName"`
	Method        string   `json:"method"`
	Args          []string `json:"args"`
}

// ContractQueryRequest 合约查询请求参数
type ContractQueryRequest struct {
	SdkConfig     string   `json:"sdkConfig"`
	IsGm          bool     `yaml:"isGM"`
	IsSM3         bool     `yaml:"isSM3"`
	ChaincodeName string   `json:"ChaincodeName"`
	Method        string   `json:"method"`
	Args          []string `json:"args"` // 假设是字符串数组，后续可转为字节
}

// ContractEventSubscribeRequest 合约事件订阅请求参数
type ContractEventSubscribeRequest struct {
	SdkConfig     string `json:"sdkConfig"`
	IsGm          bool   `yaml:"isGM"`
	IsSM3         bool   `yaml:"isSM3"`
	ChaincodeName string `json:"chaincodeName"`
	EventName     string `json:"eventName"`
	ChainName     string `json:"chainName"`
	FromBlock     string `json:"fromBlock"`
	EndBlock      string `json:"endBlock"`
}

type ContractEventUnSubscribeRequest struct {
	SdkConfig   string `json:"sdkConfig"`
	IsGm        bool   `yaml:"isGM"`
	IsSM3       bool   `yaml:"isSM3"`
	SubscribeId string `json:"subscribeId"`
}

type ContractListRequest struct {
	SdkConfig     string `json:"sdkConfig"`
	IsGm          bool   `yaml:"isGM"`
	IsSM3         bool   `yaml:"isSM3"`
	ChaincodeName string `json:"chaincodeName"`
}

type GetBlockRequest struct {
	SdkConfig   string `json:"sdkConfig"`
	IsGm        bool   `yaml:"isGM"`
	IsSM3       bool   `yaml:"isSM3"`
	BlockNumber string `json:"blockNumber"`
	OnlyHeader  bool   `json:"onlyHeader"`
}

type GetTxRequest struct {
	SdkConfig   string `json:"sdkConfig"`
	IsGm        bool   `yaml:"isGM"`
	IsSM3       bool   `yaml:"isSM3"`
	TxId        string `json:"txId"`
	BlockNumber uint64 `json:"blockNumber"`
	IsVerified  bool   `json:"isVerified"`
}

type EventRes struct {
	BlockHeight   uint64   `json:"block_height"`
	ChainId       string   `json:"chain_id"`
	TxId          string   `json:"tx_id"`
	Path          string   `json:"path"`
	EventData     []string `json:"event_data"`
	ChaincodeName string   `json:"chaincode_name"`
}

type Event struct {
	Action        string `json:"action"`        // 事件动作，例如 "setEvidence"
	Creator       string `json:"creator"`       // 创建者
	EvidenceBytes string `json:"evidenceBytes"` // 证据字节数据
}
