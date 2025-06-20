package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/qctc/fabric2-api-server/service"
	"io"
	"log"
)

func InitializeSDKBySdkId(sdkConfig string) (error, *service.Fabric2Service) {
	// 获取全局配置中的 Fabric 网络信息
	//计算sdkConfig的md5
	sdkId := fmt.Sprintf("%x", MD5Hash(sdkConfig))
	if sdk, ok := service.Fabric2ServicePool[sdkId]; ok {
		log.Printf("SDK already initialized  wait: %s", sdkId)
		return nil, sdk
	}
	err := service.InitFabric2Service(sdkConfig, sdkId)
	if err != nil {
		return err, nil
	}
	// 从池中获取刚刚初始化的实例
	sdk, ok := service.Fabric2ServicePool[sdkId]
	if !ok {
		return errors.New("failed to retrieve initialized service from pool"), nil
	}

	return nil, sdk
}

func MD5Hash(input string) string {
	hasher := md5.New()
	_, _ = io.WriteString(hasher, input)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
