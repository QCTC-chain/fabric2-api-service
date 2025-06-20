package controller

import (
	"fmt"
	"github.com/qctc/fabric2-api-server/service"
	"github.com/qctc/fabric2-api-server/utils"
	"net/http"
)

func TestConnection(w http.ResponseWriter, r *http.Request) {
	fabricName := r.URL.Query().Get("chainName")
	channelId := r.URL.Query().Get("channelId")

	// 使用已有方法测试连接
	sdk := service.Fabric2ServicePool[fabricName]
	if _, err := sdk.GetContractList(channelId); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Blockchain connection test failed", err)
		return
	}

	utils.Success(w, fmt.Sprintf("Successfully connected to blockchain"))
}
