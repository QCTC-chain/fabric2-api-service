package controller

import (
	"encoding/json"
	"fmt"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/utils"
	"log"
	"net/http"
)

func TestConnection(w http.ResponseWriter, r *http.Request) {
	// 使用已有方法测试连接
	log.Printf("test connection start --------")
	var req define.SdkConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
	if err != nil {
		fmt.Printf("sdk Initialize error --------%s", err)
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}
	if _, err := sdk.GetContractList(); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Blockchain connection test failed", err)
		return
	}

	utils.Success(w, fmt.Sprintf("Successfully connected to blockchain"))
}
