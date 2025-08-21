package utils

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/qctc/fabric2-api-server/define"
)

func UnmarshalBlock(txBytes []byte) ([]define.EventData, error) {
	tx := &common.Envelope{}
	if err := proto.Unmarshal(txBytes, tx); err != nil {
		return nil, err
	}

	// 解析交易内容
	txPayload := &common.Payload{}
	if err := proto.Unmarshal(tx.Payload, txPayload); err != nil {
		return nil, err
	}

	// 获取链码事件
	txBody := &pb.Transaction{}
	if err := proto.Unmarshal(txPayload.Data, txBody); err != nil {
		return nil, err
	}

	var eventData []define.EventData
	for _, action := range txBody.Actions {
		chaincodeActionPayload := &pb.ChaincodeActionPayload{}
		if err := proto.Unmarshal(action.Payload, chaincodeActionPayload); err != nil {
			return nil, err
		}
		proposalResponsePayload := &pb.ProposalResponsePayload{}
		if err := proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload); err != nil {
			return nil, err
		}

		chaincodeAction := &pb.ChaincodeAction{}
		if err := proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction); err != nil {
			return nil, err
		}

		ccEvent := &pb.ChaincodeEvent{}
		if err := proto.Unmarshal(chaincodeAction.Events, ccEvent); err != nil {
			return nil, err
		}
		var payload []string
		_ = json.Unmarshal(ccEvent.Payload, &payload)
		eventData = append(eventData, define.EventData{
			ChaincodeId: ccEvent.ChaincodeId,
			EventName:   ccEvent.EventName,
			Payload:     payload,
			TxId:        ccEvent.TxId,
		})

	}

	//chaincodeActionPayload := &pb.ChaincodeActionPayload{}
	//if err := proto.Unmarshal(txBody.Actions[0].Payload, chaincodeActionPayload); err != nil {
	//	return "", "", "", nil, err
	//}
	//
	//proposalResponsePayload := &pb.ProposalResponsePayload{}
	//if err := proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload); err != nil {
	//	return "", "", "", nil, err
	//}
	//
	//chaincodeAction := &pb.ChaincodeAction{}
	//if err := proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction); err != nil {
	//	return "", "", "", nil, err
	//}
	//
	//ccEvent := &pb.ChaincodeEvent{}
	//if err := proto.Unmarshal(chaincodeAction.Events, ccEvent); err != nil {
	//	return "", "", "", nil, err
	//}
	//var payload []string
	//_ = json.Unmarshal(ccEvent.Payload, &payload)

	return eventData, nil

}

func GetEventByte(block *common.Block, chainName, chaincodeName, chainId string) ([]define.EventByteData, error) {
	var eventRes define.EventRes
	var eventBytes []define.EventByteData
	eventData, err := UnmarshalBlock(block.Data.Data[0])
	if err != nil {
		return nil, err
	}
	for _, v := range eventData {
		eventRes.Path = "cross." + chainName + "." + chaincodeName
		eventRes.EventData = v.Payload
		eventRes.TxId = v.TxId
		tm, err := getTimeFromBlock(block)
		if err == nil {
			eventRes.TxTime = tm.Format(time.DateTime)
		}
		eventRes.ChaincodeName = v.ChaincodeId
		eventRes.BlockHeight = block.GetHeader().GetNumber()
		eventRes.ChainId = chainId
		eventRes.Topic = v.EventName
		eventByte, _ := json.Marshal(eventRes)
		eventBytes = append(eventBytes, define.EventByteData{
			ChaincodeId: v.ChaincodeId,
			EventName:   v.EventName,
			EventByte:   eventByte,
		})
	}
	return eventBytes, nil
}

func getTimeFromBlock(block *common.Block) (time.Time, error) {
	if block == nil || block.Data == nil || block.Data.Data == nil || len(block.Data.Data) == 0 {
		return time.Time{}, errors.New("block or block data is nil or empty")
	}

	envelopeBytes := block.Data.Data[0]
	var envelope = &common.Envelope{}
	if err := proto.Unmarshal(envelopeBytes, envelope); err != nil {
		return time.Time{}, errors.New(err.Error())
	}

	// 3. 解组为 Payload 结构体
	var payload = &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return time.Time{}, errors.New(err.Error())
	}

	if payload.Header == nil {
		return time.Time{}, errors.New("payload header is nil")
	}

	// 4. 获取 ChannelHeader
	chdr, err := getChannelHeader(payload.Header)
	if err != nil {
		return time.Time{}, err
	}

	// 5. 提取并转换时间戳
	// chdr.Timestamp 是 *timestamp.Timestamp 类型
	if chdr.Timestamp == nil {
		return time.Time{}, errors.New("channel header timestamp is nil")
	}
	blockTime := time.Unix(chdr.Timestamp.Seconds, int64(chdr.Timestamp.Nanos))
	return blockTime, nil
}

// getChannelHeader 是一个辅助函数，用于从 Payload Header 中解组出 ChannelHeader
func getChannelHeader(header *common.Header) (*common.ChannelHeader, error) {
	chdr := &common.ChannelHeader{}
	if err := proto.Unmarshal(header.ChannelHeader, chdr); err != nil {
		return nil, errors.New(err.Error())
	}
	return chdr, nil
}
