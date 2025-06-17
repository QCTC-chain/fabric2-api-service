package controller

import (
	"fmt"
	"github.com/qctc/fabric2-api-server/service"
	"github.com/qctc/fabric2-api-server/utils"
	"net/http"
)

func TestConnection(w http.ResponseWriter, r *http.Request) {
	fabricName := r.URL.Query().Get("chainName")

	err := utils.InitializeSDKByChainName(fabricName)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
		return
	}

	// 使用已有方法测试连接
	if _, err := service.GetFabric2Service().GetContractList("mychannel"); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Blockchain connection test failed", err)
		return
	}

	utils.Success(w, fmt.Sprintf("Successfully connected to blockchain"))
}
