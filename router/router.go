package router

import (
	"github.com/gorilla/mux"

	"github.com/qctc/fabric2-api-server/controller"
)

func SetUpRouter() *mux.Router {
	router := mux.NewRouter()

	// 配置相关
	//router.HandleFunc("/api/v1/config/init", controller.InitSdkConfig).Methods("POST")
	//router.HandleFunc("/api/v1/service/instantiate", controller.InstantiateService).Methods("POST")

	// 连接相关
	router.HandleFunc("/api/v1/connect/test", controller.TestConnection).Methods("GET")
	// 合约相关
	router.HandleFunc("/api/v1/contract/list", controller.GetContractList).Methods("GET")
	// 调用智能合约
	router.HandleFunc("/api/v1/contract/sendTransaction", controller.InvokeContract).Methods("POST")

	// 查询智能合约
	router.HandleFunc("/api/v1/contract/call", controller.QueryContract).Methods("POST")

	//获取合约信息
	router.HandleFunc("/api/v1/contract/info", controller.GetContractInfo).Methods("GET")

	//获取区块信息
	router.HandleFunc("/api/v1/block/info", controller.GetBlockInfo).Methods("GET")

	//获取交易信息
	router.HandleFunc("/api/v1/transaction/info", controller.GetTransactionInfo).Methods("GET")

	//订阅合约事件
	router.HandleFunc("/api/v1/contract/subscribe", controller.SubscribeContractEvent).Methods("POST")

	//取消订阅合约事件
	router.HandleFunc("/api/v1/contract/unsubscribe", controller.UnsubscribeContractEvent).Methods("POST")

	return router
}
