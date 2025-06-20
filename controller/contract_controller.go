package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/service"
	"github.com/qctc/fabric2-api-server/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetContractList(w http.ResponseWriter, r *http.Request) {
	chainName := r.URL.Query().Get("chainName")
	channelId := r.URL.Query().Get("channelId")
	mspId := r.URL.Query().Get("mspId")
	userName := r.URL.Query().Get("userName")
	pathId := r.URL.Query().Get("pathId")
	peerList := r.URL.Query().Get("peers")
	if mspId == "" || userName == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	var peers []string
	if peerList != "" {
		peers = strings.Split(peerList, ",")
	}
	var userConfig = define.UserConfigRequest{
		MspId:    mspId,
		PathId:   pathId,
		Peers:    peers,
		UserName: userName,
	}
	isOldSDK, err := utils.UpdateUserInConfig(userConfig, chainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(chainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(chainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	if channelId == "" {
		utils.BadRequest(w, "channelId is required")
		return
	}

	contracts, err := fabric2Service.GetContractList(channelId)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, contracts)
}

func GetContractInfo(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channelId")
	chaincodeName := r.URL.Query().Get("chaincodeName")
	chainName := r.URL.Query().Get("chainName")
	if channelID == "" || chaincodeName == "" {
		utils.BadRequest(w, "channelId and chaincode are required")
		return
	}

	mspId := r.URL.Query().Get("mspId")
	userName := r.URL.Query().Get("userName")
	pathId := r.URL.Query().Get("pathId")
	peerList := r.URL.Query().Get("peers")
	if mspId == "" || userName == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	var peers []string
	if peerList != "" {
		peers = strings.Split(peerList, ",")
	}
	var userConfig = define.UserConfigRequest{
		MspId:    mspId,
		PathId:   pathId,
		Peers:    peers,
		UserName: userName,
	}
	isOldSDK, err := utils.UpdateUserInConfig(userConfig, chainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(chainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(chainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	info, err := fabric2Service.GetContractInfo(channelID, chaincodeName)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, info)
}

func InvokeContract(w http.ResponseWriter, r *http.Request) {
	var req define.ContractInvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.ChannelID == "" || req.ChaincodeName == "" || req.Method == "" {
		utils.BadRequest(w, "channel_id, chaincode and function are required")
		return
	}
	if req.UserConfig.UserName == "" || req.UserConfig.MspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	isOldSDK, err := utils.UpdateUserInConfig(req.UserConfig, req.ChainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(req.ChainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(req.ChainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}
	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}
	resp, err := fabric2Service.InvokeContract(req.ChannelID, req.ChaincodeName, req.Method, args)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, map[string]interface{}{
		"payload": string(resp),
	})
}

func QueryContract(w http.ResponseWriter, r *http.Request) {
	var req define.ContractQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.ChannelID == "" || req.ChaincodeID == "" || req.Method == "" {
		utils.BadRequest(w, "channel_id, chaincode and function are required")
		return
	}
	if req.UserConfig.UserName == "" || req.UserConfig.MspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	isOldSDK, err := utils.UpdateUserInConfig(req.UserConfig, req.ChainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(req.ChainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(req.ChainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}

	resp, err := fabric2Service.InvokeContract(req.ChannelID, req.ChaincodeID, req.Method, args)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, map[string]interface{}{
		"payload": string(resp),
	})
}

func SubscribeContractEvent(w http.ResponseWriter, r *http.Request) {
	var req define.ContractEventSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.ChannelID == "" || req.ChaincodeID == "" {
		utils.BadRequest(w, "channelId, chaincode are required")
		return
	}
	if req.UserConfig.UserName == "" || req.UserConfig.MspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	isOldSDK, err := utils.UpdateUserInConfig(req.UserConfig, req.ChainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(req.ChainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(req.ChainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	regID, eventCh, err := fabric2Service.SubscribeEvent(req.ChannelID, req.ChaincodeID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	// 构造唯一 key
	key := fmt.Sprintf("%s:%s:%s", req.ChainName, req.ChannelID, req.ChaincodeID)

	// 存储 regID
	define.SubscriptionMutex.Lock()
	define.EventSubscriptions[key] = regID
	define.SubscriptionMutex.Unlock()

	utils.Success(w, map[string]interface{}{
		"payload": "subscribed successfully",
	})
	// 模拟简单事件监听逻辑
	go func() {
		for {
			select {
			case event := <-eventCh:
				message := &primitive.Message{
					Topic: define.GlobalConfig.MQ.Topic,
					Body:  []byte(fmt.Sprintf("%v", event)), // 可以根据实际格式序列化 event
				}
				_, err := define.GlobalProducer.SendSync(context.Background(), message)
				if err != nil {
					log.Printf("Failed to send message to RocketMQ: %v", err)
				} else {
					log.Printf("Event sent to RocketMQ: %v", event)
				}
			}
		}
	}()

}

func UnsubscribeContractEvent(w http.ResponseWriter, r *http.Request) {
	var req define.ContractEventUnSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.ChannelID == "" || req.ChaincodeID == "" {
		utils.BadRequest(w, "channelId, chaincode are required")
		return
	}
	if req.UserConfig.UserName == "" || req.UserConfig.MspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	isOldSDK, err := utils.UpdateUserInConfig(req.UserConfig, req.ChainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(req.ChainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(req.ChainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	// 构造 key
	key := fmt.Sprintf("%s:%s:%s", req.ChainName, req.ChannelID, req.ChaincodeID)

	// 获取 regID
	define.SubscriptionMutex.Lock()
	regID, exists := define.EventSubscriptions[key]
	define.SubscriptionMutex.RUnlock()
	if !exists {
		utils.BadRequest(w, "subscription not found")
		return
	}
	err = fabric2Service.UnsubscribeEvent(req.ChannelID, req.ChaincodeID, regID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	// 删除缓存
	define.SubscriptionMutex.Lock()
	delete(define.EventSubscriptions, key)
	define.SubscriptionMutex.RUnlock()

	utils.Success(w, map[string]interface{}{
		"message": "unsubscribed successfully",
	})

	define.GlobalProducer.Shutdown()
}

func GetBlockInfo(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channelId")
	blockNumberStr := r.URL.Query().Get("blockNumber")
	chainName := r.URL.Query().Get("chainName")

	if channelID == "" || blockNumberStr == "" {
		utils.BadRequest(w, "channel_id and block_number are required")
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, "invalid block_number format")
		return
	}
	mspId := r.URL.Query().Get("mspId")
	userName := r.URL.Query().Get("userName")
	pathId := r.URL.Query().Get("pathId")
	peerList := r.URL.Query().Get("peers")
	if userName == "" || mspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	var peers []string
	if peerList != "" {
		peers = strings.Split(peerList, ",")
	}
	var userConfig = define.UserConfigRequest{
		MspId:    mspId,
		PathId:   pathId,
		Peers:    peers,
		UserName: userName,
	}
	isOldSDK, err := utils.UpdateUserInConfig(userConfig, chainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(chainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(chainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	block, err := fabric2Service.GetBlockInfo(channelID, blockNumber)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, block)
}

func GetTransactionInfo(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channelId")
	txID := r.URL.Query().Get("txId")
	chainName := r.URL.Query().Get("chainName")

	if channelID == "" || txID == "" {
		utils.BadRequest(w, "channelId and txId are required")
		return
	}
	mspId := r.URL.Query().Get("mspId")
	userName := r.URL.Query().Get("userName")
	pathId := r.URL.Query().Get("pathId")
	peerList := r.URL.Query().Get("peers")
	if userName == "" || mspId == "" {
		utils.BadRequest(w, "userName and mspId are required")
		return
	}
	var peers []string
	if peerList != "" {
		peers = strings.Split(peerList, ",")
	}
	var userConfig = define.UserConfigRequest{
		MspId:    mspId,
		PathId:   pathId,
		Peers:    peers,
		UserName: userName,
	}
	isOldSDK, err := utils.UpdateUserInConfig(userConfig, chainName)
	if err != nil {
		utils.BadRequest(w, "UpdateUserInConfig error")
		return
	}
	if !isOldSDK {
		err = utils.InitializeSDKByChainName(chainName)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to initialize SDK", err)
			return
		}
	}

	fabric2Service := service.GetFabric2Service(chainName)
	if fabric2Service == nil {
		utils.BadRequest(w, "Fabric2Service is not initialized")
		return
	}

	tx, err := fabric2Service.GetTransactionInfo(channelID, txID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, tx)
}
