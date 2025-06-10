package controller

import (
	"io"
	"log"
	"net/http"

	"github.com/qctc/fabric2-api-server/service"
	"github.com/qctc/fabric2-api-server/utils"
)

func InitSdkConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}
	utils.Success(w, string(body))
}

func InstantiateService(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	if err := service.InitFabric2Service("./config/config.yaml"); err != nil {
		log.Fatalf("Failed to initialize Fabric2Service: %v", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to initialize Fabric2Service", err)
		return
	}

	utils.Success(w, string(body))
}
