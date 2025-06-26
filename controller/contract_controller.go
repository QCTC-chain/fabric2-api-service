package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/hyperledger/fabric-protos-go/common"
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}
	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}
	//var args [][]byte
	//arg, _ := json.Marshal(req.Args)
	//args = append(args, arg)
	resp, txId, err := sdk.InvokeContract(req.ChaincodeName, req.Method, args)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	height, err := sdk.GetBlockByTxID(string(txId))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	// 解析 TransactionEnvelope.Payload（是一个 []byte）

	utils.Success(w, map[string]interface{}{
		"payload": string(resp),
		"txHash":  string(txId),
		"height":  height,
	})
}

func QueryContract(w http.ResponseWriter, r *http.Request) {
	var req define.ContractQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	// 将 args 转为 [][]byte
	args := make([][]byte, len(req.Args))
	for i, arg := range req.Args {
		args[i] = []byte(arg)
	}

	resp, txId, err := sdk.QueryContract(req.ChaincodeName, req.Method, args)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	block, err := sdk.GetBlockInfo("latest")
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	// 解析 TransactionEnvelope.Payload（是一个 []byte）

	utils.Success(w, map[string]interface{}{
		"payload": string(resp),
		"txHash":  string(txId),
		"height":  block.GetHeader().GetNumber(),
	})
}

func SubscribeContractEvent(w http.ResponseWriter, r *http.Request) {
	var req define.ContractEventSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	regID, eventCh, chainId, err := sdk.SubscribeEvent(req.ChaincodeName, req.EventName)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	sdkId := fmt.Sprintf("%x", utils.MD5Hash(req.SdkConfig))
	const subscriptionKeyFormat = "%s:%s:%s"
	key := fmt.Sprintf(subscriptionKeyFormat, sdkId, req.ChaincodeName, req.EventName)

	// 使用 sync.Map 替代 Mutex + map
	define.SubscriptionMutex.Lock()
	define.EventSubscriptions[key] = regID
	define.SubscriptionMutex.Unlock()

	log.Printf("Subscribed to event: %s on chaincode: %s with key: %s", req.EventName, req.ChaincodeName, key)

	// 获取历史区块并发送事件
	blocks, err := sdk.GetBlocks(req.FromBlock)
	if err != nil {
		log.Printf("Failed to fetch blocks starting from %d: %v", req.FromBlock, err)
		utils.InternalServerError(w, err)
		return
	}
	for _, block := range blocks {
		sendEventToRocketMQ(req.EventName, req.ChaincodeName, req.ChainName, block, chainId, define.GlobalProducer)
	}
	utils.Success(w, map[string]interface{}{
		"subscribeId": key,
	})
	// 启动监听协程，并通过 context 管理生命周期
	ctx, cancel := context.WithCancel(context.Background())
	define.SubscriptionContext.Store(key, cancel) // 存储 cancel 用于后续取消

	go func(ctx context.Context) {
		defer log.Printf("Stopped listener for subscription: %s", key)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					continue
				}
				sendEventToRocketMQ(req.EventName, req.ChaincodeName, req.ChainName, event.Block, chainId, define.GlobalProducer)
			case <-ctx.Done():
				return
			}
		}
	}(ctx)
}
func UnsubscribeContractEvent(w http.ResponseWriter, r *http.Request) {
	var req define.ContractEventUnSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
	if err != nil {
		utils.BadRequest(w, fmt.Sprintf("sdk Initialize error %s", err))
		return
	}

	// 获取 regID
	define.SubscriptionMutex.RLocker().Lock()
	regID, exists := define.EventSubscriptions[req.SubscribeId]
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
	define.SubscriptionMutex.RLocker().Lock()
	delete(define.EventSubscriptions, req.SubscribeId)
	define.SubscriptionMutex.RUnlock()

	if cancel, ok := define.SubscriptionContext.Load(req.SubscribeId); ok {
		cancel.(context.CancelFunc)()
	}
	define.SubscriptionContext.Delete(req.SubscribeId)

	err = define.GlobalProducer.GracefulStop()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	utils.Success(w, map[string]interface{}{
		"subscribeId": req.SubscribeId,
	})

}

func GetBlockInfo(w http.ResponseWriter, r *http.Request) {
	var req define.GetBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
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
	err, sdk := utils.InitializeSDKBySdkId(req.SdkConfig, req.IsGm, req.IsSM3)
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

// 封装发送事件到 RocketMQ 的逻辑
func sendEventToRocketMQ(eventName, chaincodeName, chainName string, block *common.Block, chainId string, producer golang.Producer) {
	eventNameRes, chaincodeID, eventByte, err := utils.GetEventByte(block, chainName, chaincodeName, chainId)
	if err != nil {
		log.Printf("Failed to get event byte: %v", err)
		return
	}
	if eventNameRes == eventName && chaincodeID == chaincodeName {
		message := &golang.Message{
			Topic: define.GlobalConfig.MQ.Topic,
			Body:  eventByte,
		}
		_, err := producer.Send(context.TODO(), message)
		if err != nil {
			log.Printf("Failed to send message to RocketMQ: %v", err)
		} else {
			log.Printf("Event sent to RocketMQ")
		}
	}
}
