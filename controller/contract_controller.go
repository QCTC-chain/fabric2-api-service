package controller

import (
	"net/http"

	"github.com/qctc/fabric2-api-server/service"
	"github.com/qctc/fabric2-api-server/utils"
)

func GetContractList(w http.ResponseWriter, r *http.Request) {
	fabric2Service := service.GetFabric2Service()
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	channelID := r.URL.Query().Get("channel_id")

	if channelID == "" {
		utils.BadRequest(w, "channel_id is required")
		return
	}

	contracts, err := fabric2Service.GetContractList(channelID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, contracts)
}
