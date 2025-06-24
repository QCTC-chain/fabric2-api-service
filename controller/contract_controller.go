package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/utils"
	"log"
	"net/http"
)

func GetContractList(w http.ResponseWriter, r *http.Request) {
	var req define.SdkConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	contracts, err := sdk.GetContractList()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, contracts)
}

func GetContractInfo(w http.ResponseWriter, r *http.Request) {
	var req define.ContractListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	info, err := sdk.GetContractInfo(req.ChaincodeName)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}
	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}
	resp, err := sdk.InvokeContract(req.ChaincodeName, req.Method, args)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}

	resp, err := sdk.QueryContract(req.ChaincodeName, req.Method, args)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	regID, eventCh, err := sdk.SubscribeEvent(req.ChaincodeName, req.EventName)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	sdkId := fmt.Sprintf("%x", utils.MD5Hash(req.SdkConfig))
	// 构造唯一 key
	key := fmt.Sprintf("%s:%s:%s", sdkId, req.ChaincodeName, req.EventName)

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
	var req define.ContractEventSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	sdkId := fmt.Sprintf("%x", utils.MD5Hash(req.SdkConfig))
	// 构造唯一 key
	key := fmt.Sprintf("%s:%s:%s", sdkId, req.ChaincodeName, req.EventName)
	// 构造 key

	// 获取 regID
	define.SubscriptionMutex.Lock()
	regID, exists := define.EventSubscriptions[key]
	define.SubscriptionMutex.RUnlock()
	if !exists {
		utils.BadRequest(w, "subscription not found")
		return
	}
	err = sdk.UnsubscribeEvent(regID)
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
	var req define.GetBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	block, err := sdk.GetBlockInfo(req.BlockNumber)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	blockInfo := map[string]interface{}{
		"blockNumber":  block.Header.Number,
		"previousHash": block.Header.PreviousHash,
		"dataHash":     block.Header.DataHash,
	}

	utils.Success(w, blockInfo)
}

func GetTransactionInfo(w http.ResponseWriter, r *http.Request) {
	var req define.GetTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	tx, err := sdk.GetTransactionInfo(req.TxId)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, tx)
}
