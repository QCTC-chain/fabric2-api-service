package service

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestFabric2(t *testing.T) {
	fabricSDK, _ := fabsdk.New(config.FromFile(""), fabsdk.WithGMTLS(true), fabsdk.WithTxTimeStamp(false))
	ctx := fabricSDK.ChannelContext("", fabsdk.WithOrg("organizationName"), fabsdk.WithUser(""))
	client, _ := channel.New(ctx)
	req := channel.Request{}
	res, _ := client.Execute(req, channel.WithTargetEndpoints([]string{"peer"}...), channel.WithRetry(retry.DefaultChannelOpts))
	_ = res.Responses[0].Response.Payload
}
