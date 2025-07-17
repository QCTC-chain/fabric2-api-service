/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

// InfraProvider represents the default implementation of Fabric objects.
type InfraProvider struct {
	providerContext   context.Providers
	commManager       *comm.CachingConnector
	isGMTLS           bool //add by liuhy for gm tls
	enableTxTimeStamp bool // jzk, tx with timestamp, for fabric 1.4.8-enhanced
	isSm3             bool
}

// New creates a InfraProvider enabling access to core Fabric objects and functionality.
func New(config fab.EndpointConfig) *InfraProvider {
	idleTime := config.Timeout(fab.ConnectionIdle)
	sweepTime := config.Timeout(fab.CacheSweepInterval)

	return &InfraProvider{
		commManager: comm.NewCachingConnector(sweepTime, idleTime),
	}
}

func (f *InfraProvider) IsSm3() bool {
	return f.isSm3
}

// Initialize sets the provider context
func (f *InfraProvider) Initialize(isGMTLS bool, enableTxTimeStamp bool, isSM3 bool, providers context.Providers) error { // jzk, tx with timestamp, for fabric 1.4.8-enhanced
	f.providerContext = providers
	f.isGMTLS = isGMTLS
	f.enableTxTimeStamp = enableTxTimeStamp
	f.isSm3 = isSM3
	return nil
}

// Close frees resources and caches.
func (f *InfraProvider) Close() {
	logger.Debug("Closing comm manager...")
	f.commManager.Close()
}

// CommManager provides comm support such as GRPC onnections
func (f *InfraProvider) CommManager() fab.CommManager {
	return f.commManager
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *InfraProvider) CreatePeerFromConfig(isGMTLS bool, peerCfg *fab.NetworkPeer) (fab.Peer, error) {
	//add by liuhy for gm tls
	if isGMTLS {
		return peerImpl.New(f.providerContext.EndpointConfig(), peerImpl.FromPeerConfig(peerCfg), peerImpl.WithInGMTLS())
	}
	return peerImpl.New(f.providerContext.EndpointConfig(), peerImpl.FromPeerConfig(peerCfg))
}

// CreateOrdererFromConfig creates a default implementation of Orderer based on configuration.
func (f *InfraProvider) CreateOrdererFromConfig(isGMTLS bool, cfg *fab.OrdererConfig) (fab.Orderer, error) {
	//add by liuhy for gm tls
	if isGMTLS {
		newOrderer, err := orderer.New(f.providerContext.EndpointConfig(), orderer.FromOrdererConfig(cfg), orderer.WithInGMTLS())
		if err != nil {
			return nil, errors.WithMessage(err, "creating orderer failed")
		}
		return newOrderer, nil
	}

	newOrderer, err := orderer.New(f.providerContext.EndpointConfig(), orderer.FromOrdererConfig(cfg))
	if err != nil {
		return nil, errors.WithMessage(err, "creating orderer failed")
	}
	return newOrderer, nil
}

func (f *InfraProvider) IsGMTLS() bool {
	return f.isGMTLS
}

// jzk, tx with timestamp, for fabric 1.4.8-enhanced
func (f *InfraProvider) EnableTxTimeStamp() bool {
	return f.enableTxTimeStamp
}
