package utils

import (
	"errors"
	"fmt"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/service"
)

func InitializeSDKByChainName(fabricName string) error {
	// 获取全局配置中的 Fabric 网络信息
	fabricCfg := define.GlobalConfig.Fabric

	network, exists := fabricCfg[fabricName]
	if !exists {
		return errors.New(fmt.Sprintf("%s is not exists", fabricName))
	}
	err := service.InitFabric2Service(network.ConfigFilePath, fabricName)
	if err != nil {
		return err
	}
	return nil
}
