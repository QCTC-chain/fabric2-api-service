package utils

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/qctc/fabric2-api-server/define"
)

func UnmarshalBlock(txBytes []byte) (string, string, string, []string, error) {
	tx := &common.Envelope{}
	if err := proto.Unmarshal(txBytes, tx); err != nil {
		return "", "", "", nil, err
	}

	// 解析交易内容
	txPayload := &common.Payload{}
	if err := proto.Unmarshal(tx.Payload, txPayload); err != nil {
		return "", "", "", nil, err
	}

	// 获取链码事件
	txBody := &pb.Transaction{}
	if err := proto.Unmarshal(txPayload.Data, txBody); err != nil {
		return "", "", "", nil, err
	}

	chaincodeActionPayload := &pb.ChaincodeActionPayload{}
	if err := proto.Unmarshal(txBody.Actions[0].Payload, chaincodeActionPayload); err != nil {
		return "", "", "", nil, err
	}

	proposalResponsePayload := &pb.ProposalResponsePayload{}
	if err := proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload); err != nil {
		return "", "", "", nil, err
	}

	chaincodeAction := &pb.ChaincodeAction{}
	if err := proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction); err != nil {
		return "", "", "", nil, err
	}

	ccEvent := &pb.ChaincodeEvent{}
	if err := proto.Unmarshal(chaincodeAction.Events, ccEvent); err != nil {
		return "", "", "", nil, err
	}
	var payload []string
	_ = json.Unmarshal(ccEvent.Payload, &payload)

	return ccEvent.EventName, ccEvent.TxId, ccEvent.ChaincodeId, payload, nil

}

func GetEventByte(block *common.Block, chainName, chaincodeName, chainId string) (string, string, []byte, error) {
	var eventRes define.EventRes
	eventName, txID, chaincodeID, eventPayload, err := UnmarshalBlock(block.Data.Data[0])
	if err != nil {
		return "", "", nil, err
	}
	eventRes.Path = "cross." + chainName + "." + chaincodeName
	eventRes.EventData = eventPayload
	eventRes.TxId = txID
	eventRes.ChaincodeName = chaincodeID
	eventRes.BlockHeight = block.GetHeader().GetNumber()
	eventRes.ChainId = chainId

	eventByte, _ := json.Marshal(eventRes)
	return eventName, chaincodeID, eventByte, nil
}
