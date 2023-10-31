package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/axm"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/eth"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/eth/filters"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/net"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc/namespaces/web3"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// RPC namespaces and API version
const (
	Web3Namespace   = "web3"
	EthNamespace    = "eth"
	NetNamespace    = "net"
	AxmNamespace    = "axm"
	TxPoolNamespace = "txpool"

	apiVersion = "1.0"
)

// GetAPIs returns the list of all APIs from the Ethereum namespaces
func GetAPIs(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) ([]rpc.API, error) {
	var apis []rpc.API

	apis = append(apis,
		rpc.API{
			Namespace: AxmNamespace,
			Version:   apiVersion,
			Service:   axm.NewAxmAPI(rep, api, logger),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   eth.NewBlockChainAPI(rep, api, logger),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   eth.NewAxiomAPI(rep, api, logger),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   filters.NewAPI(rep, api, logger),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   eth.NewTransactionAPI(rep, api, logger),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: Web3Namespace,
			Version:   apiVersion,
			Service:   web3.NewAPI(),
			Public:    true,
		},
	)

	apis = append(apis,
		rpc.API{
			Namespace: NetNamespace,
			Version:   apiVersion,
			Service:   net.NewAPI(rep),
			Public:    true,
		},
	)
	apis = append(apis,
		rpc.API{
			Namespace: TxPoolNamespace,
			Version:   apiVersion,
			Service:   axm.NewTxPoolAPI(rep, api, logger),
			Public:    true,
		},
	)

	return apis, nil
}
