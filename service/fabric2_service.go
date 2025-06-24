package service

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/qctc/fabric2-api-server/model/vo"
)

type Fabric2Service struct {
	sdk *fabsdk.FabricSDK
}

var fabric2ServiceInstance *Fabric2Service
var Fabric2ServicePool map[string]*Fabric2Service

func InitFabric2Service(configString string, sdkId string) error {
	sdk, err := fabsdk.New(
		//config.FromFile(configPath),
		config.FromRaw([]byte(configString), "yaml"),
		fabsdk.WithGMTLS(false),
		fabsdk.WithTxTimeStamp(false))
	if err != nil {
		return err
	}
	fabric2ServiceInstance = &Fabric2Service{sdk: sdk}
	if Fabric2ServicePool == nil {
		Fabric2ServicePool = make(map[string]*Fabric2Service)
	}
	Fabric2ServicePool[sdkId] = &Fabric2Service{sdk: sdk}
	return nil
}

func GetFabric2Service(chainName string) *Fabric2Service {
	return Fabric2ServicePool[chainName]
}

func (s *Fabric2Service) getOrgName() (string, error) {
	sdkConfig, err := s.sdk.Config()
	if err != nil {
		return "", err
	}

	clientConfig, bok := sdkConfig.Lookup("client")
	if !bok {
		return "", errors.New("client configuration not found")
	}

	clientConfigMap := clientConfig.(map[string]interface{})
	orgName := clientConfigMap["organization"].(string)

	return orgName, nil
}

func (s *Fabric2Service) getOrgAdmin(orgName string) (string, error) {
	sdkConfig, err := s.sdk.Config()
	if err != nil {
		return "", err
	}

	organizations, bok := sdkConfig.Lookup("organizations")
	if !bok {
		return "", errors.New("organizations configuration not found")
	}

	orgsMap := organizations.(map[string]interface{})
	usersMap := orgsMap[orgName].(map[string]interface{})["users"].(map[string]interface{})
	var orgAdmin string
	for userName := range usersMap {
		orgAdmin = userName
		break
	}

	return orgAdmin, nil
}

func (s *Fabric2Service) getChannelID() (string, error) {
	sdkConfig, err := s.sdk.Config()
	if err != nil {
		return "", err
	}

	// 查找 "channels" 节点
	channelsSection, ok := sdkConfig.Lookup("channels")
	if !ok {
		return "", errors.New("channels configuration not found")
	}

	channelsMap := channelsSection.(map[string]interface{})

	// 遍历找到第一个 channel 名称
	for channelID := range channelsMap {
		return channelID, nil
	}

	return "", errors.New("no channel found in configuration")
}

func (s *Fabric2Service) getPeers() ([]string, error) {
	sdkConfig, err := s.sdk.Config()
	if err != nil {
		return nil, err
	}

	// 查找 "peers" 节点
	peersSection, ok := sdkConfig.Lookup("peers")
	if !ok {
		return nil, errors.New("peers configuration not found")
	}

	peersMap := peersSection.(map[string]interface{})

	// 提取所有 Peer 名称
	var peerNames []string
	for peerName := range peersMap {
		peerNames = append(peerNames, peerName)
	}

	return peerNames, nil
}

func (s *Fabric2Service) GetContractList() ([]vo.ContractVO, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	// 创建管理用户的上下文
	orgAdmin, err := s.getOrgAdmin(orgName)

	if err != nil {
		return nil, err
	}

	//获取peer
	peers, err := s.getPeers()
	if err != nil {
		return nil, err
	}
	resourceManagerContext := s.sdk.Context(
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName))
	// 创建 resmgmt client
	resMgmtClient, err := resmgmt.New(resourceManagerContext)
	if err != nil {
		return nil, err
	}
	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}
	// 获取某个通道上已部署的 chaincode 列表
	installedCC, err := resMgmtClient.LifecycleQueryCommittedCC(channelID, resmgmt.LifecycleQueryCommittedCCRequest{}, resmgmt.WithTargetEndpoints(peers...))
	if err != nil {
		log.Printf("Failed to query committed chaincodes: %v", err)
		return nil, err
	}
	contractList := make([]vo.ContractVO, 0)
	for _, cc := range installedCC {
		contractList = append(contractList, vo.ContractVO{
			Name:     cc.Name,
			Version:  cc.Version,
			Sequence: cc.Sequence,
		})
	}

	return contractList, nil
}

// GetContractInfo 获取合约信息
func (s *Fabric2Service) GetContractInfo(chaincodeName string) (map[string]interface{}, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}

	channelContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)

	channelClient, err := channel.New(channelContext)
	if err != nil {
		return nil, err
	}

	// 调用链码查询接口（需要链码实现GetMetadata方法）
	response, err := channelClient.Query(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         "GetMetadata",
	})
	if err != nil {
		return nil, err
	}

	// 解析响应数据（根据实际链码响应结构调整）
	var metadata map[string]interface{}
	err = json.Unmarshal(response.Payload, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// SubscribeEvent 订阅合约事件
func (s *Fabric2Service) SubscribeEvent(chaincodeName string, eventName string) (fab.Registration, <-chan *fab.CCEvent, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, nil, err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, nil, err
	}

	eventContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)

	eventClient, err := event.New(eventContext)
	if err != nil {
		return nil, nil, err
	}

	// 注册事件监听
	req, eventCh, err := eventClient.RegisterChaincodeEvent(chaincodeName, eventName)
	if err != nil {
		return nil, nil, err
	}

	// 返回事件通道，调用方需要负责管理注册和清理
	return req, eventCh, nil
}

// InvokeContract 执行合约调用
func (s *Fabric2Service) InvokeContract(chaincodeName, function string, args [][]byte) ([]byte, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}

	channelContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)
	channelClient, err := channel.New(channelContext)
	if err != nil {
		return nil, err
	}
	// 执行链码调用
	response, err := channelClient.Execute(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         function,
		Args:        args,
	})
	if err != nil {
		log.Printf("execute is error %s", err)
		return nil, err
	}

	return response.Payload, nil
}

// QueryContract 查询合约调用
func (s *Fabric2Service) QueryContract(chaincodeName, function string, args [][]byte) ([]byte, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}

	channelContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)
	channelClient, err := channel.New(channelContext)
	if err != nil {
		return nil, err
	}
	// 执行链码调用
	response, err := channelClient.Query(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         function,
		Args:        args,
	})
	if err != nil {
		log.Printf("execute is error %s", err)
		return nil, err
	}

	return response.Payload, nil
}

// GetBlockInfo 获取区块信息
func (s *Fabric2Service) GetBlockInfo(blockNumber string) (*common.Block, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}

	ledgerContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)

	ledgerClient, err := ledger.New(ledgerContext)
	if err != nil {
		return nil, err
	}

	var queryBlockNumber uint64
	if blockNumber == "latest" {
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			return nil, err
		}

		queryBlockNumber = info.BCI.Height - 1
	} else {
		queryBlockNumber, err = strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	// 查询指定区块信息
	block, err := ledgerClient.QueryBlock(queryBlockNumber)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// UnsubscribeEvent 取消事件订阅（需要传递注册ID）
func (s *Fabric2Service) UnsubscribeEvent(regID fab.Registration) error {
	orgName, err := s.getOrgName()
	if err != nil {
		return err
	}

	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return err
	}
	channelID, err := s.getChannelID()
	if err != nil {
		return err
	}

	eventContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)

	eventClient, err := event.New(eventContext)
	if err != nil {
		return err
	}

	// 取消事件注册
	eventClient.Unregister(regID)
	return nil
}

func (s *Fabric2Service) TestConnection() error {
	orgName, err := s.getOrgName()
	if err != nil {
		log.Println("Failed to get organization name")
		return err
	}
	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		log.Println("Failed to get organization admin")
		return err
	}

	// 创建管理用户上下文
	ctx := s.sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	// 创建资源管理客户端（用于与排序节点通信）
	resMgmtClient, err := resmgmt.New(ctx)
	if err != nil {
		log.Println("Failed to create resource management client")
		return err
	}

	// 查询通道信息（实际发送请求，验证是否能正常通信）
	channels, err := resMgmtClient.QueryChannels()
	if err != nil {
		log.Println("Failed to query channels")
		return err
	}
	if len(channels.Channels) > 0 {
		return nil
	}

	return nil
}

// GetTransactionInfo 获取指定交易的详细信息
func (s *Fabric2Service) GetTransactionInfo(txID string) (*pb.ProcessedTransaction, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}
	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}

	channelID, err := s.getChannelID()
	if err != nil {
		return nil, err
	}

	ledgerContext := s.sdk.ChannelContext(
		channelID,
		fabsdk.WithUser(orgAdmin),
		fabsdk.WithOrg(orgName),
	)

	ledgerClient, err := ledger.New(ledgerContext)
	if err != nil {
		return nil, err
	}

	// 查询交易详情
	tx, err := ledgerClient.QueryTransaction(fab.TransactionID(txID))
	if err != nil {
		return nil, err
	}

	return tx, nil
}
