package service

import (
	"errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/qctc/fabric2-api-server/model/vo"
)

type Fabric2Service struct {
	sdk *fabsdk.FabricSDK
}

var fabric2ServiceInstance *Fabric2Service

func InitFabric2Service(configPath string) error {
	sdk, err := fabsdk.New(
		config.FromFile(configPath),
		fabsdk.WithGMTLS(true),
		fabsdk.WithTxTimeStamp(false))
	if err != nil {
		return err
	}
	fabric2ServiceInstance = &Fabric2Service{sdk: sdk}
	return nil
}

func GetFabric2Service() *Fabric2Service {
	return fabric2ServiceInstance
}

func (s *Fabric2Service) getOrgName() (string, error) {
	config, err := s.sdk.Config()
	if err != nil {
		return "", err
	}

	clientConfig, bok := config.Lookup("client")
	if !bok {
		return "", errors.New("client configuration not found")
	}

	clientConfigMap := clientConfig.(map[string]interface{})
	orgName := clientConfigMap["organization"].(string)

	return orgName, nil
}

func (s *Fabric2Service) getOrgAdmin(orgName string) (string, error) {
	config, err := s.sdk.Config()
	if err != nil {
		return "", err
	}

	organizations, bok := config.Lookup("organizations")
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

func (s *Fabric2Service) GetContractList(channelID string) ([]vo.ContractVO, error) {
	orgName, err := s.getOrgName()
	if err != nil {
		return nil, err
	}

	// 创建管理用户的上下文
	orgAdmin, err := s.getOrgAdmin(orgName)
	if err != nil {
		return nil, err
	}
	resourceManagerContext := s.sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	// 创建 resmgmt client
	resMgmtClient, err := resmgmt.New(resourceManagerContext)
	if err != nil {
		return nil, err
	}

	// 获取某个通道上已部署的 chaincode 列表
	installedCC, err := resMgmtClient.LifecycleQueryCommittedCC(channelID, resmgmt.LifecycleQueryCommittedCCRequest{})
	if err != nil {
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
