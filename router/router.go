package router

import (
	"github.com/gorilla/mux"

	"github.com/qctc/fabric2-api-server/controller"
)

func SetUpRouter() *mux.Router {
	router := mux.NewRouter()

	// 配置相关
	router.HandleFunc("/api/v1/config/init", controller.InitSdkConfig).Methods("POST")
	router.HandleFunc("/api/v1/service/instantiate", controller.InstantiateService).Methods("POST")

	// 合约相关
	router.HandleFunc("/api/v1/contract/list", controller.GetContractList).Methods("GET")
	return router
}
