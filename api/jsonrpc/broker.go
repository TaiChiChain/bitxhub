package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/ratelimiter"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type ChainBrokerService struct {
	rep *repo.Repo

	// genesis     *repo.Genesis
	api api.CoreAPI

	server              *rpc.Server
	wsServer            *rpc.Server
	logger              logrus.FieldLogger
	rateLimiterForRead  *ratelimiter.JRateLimiter
	rateLimiterForWrite *ratelimiter.JRateLimiter

	ctx    context.Context
	cancel context.CancelFunc
}

func NewChainBrokerService(coreAPI api.CoreAPI, rep *repo.Repo) (*ChainBrokerService, error) {
	logger := loggers.Logger(loggers.API)

	config := rep.Config
	readLimiter, err := ratelimiter.NewJRateLimiterWithQuantum(config.JsonRPC.ReadLimiter.Interval.ToDuration(), config.JsonRPC.ReadLimiter.Capacity, config.JsonRPC.ReadLimiter.Quantum)
	if err != nil {
		return nil, fmt.Errorf("create read rate limiter failed: %w", err)
	}

	writeLimiter, err := ratelimiter.NewJRateLimiterWithQuantum(config.JsonRPC.WriteLimiter.Interval.ToDuration(), config.JsonRPC.WriteLimiter.Capacity, config.JsonRPC.WriteLimiter.Quantum)
	if err != nil {
		return nil, fmt.Errorf("create write rate limiter failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cbs := &ChainBrokerService{
		logger:              logger,
		rep:                 rep,
		api:                 coreAPI,
		ctx:                 ctx,
		cancel:              cancel,
		rateLimiterForRead:  readLimiter,
		rateLimiterForWrite: writeLimiter,
	}

	if err := cbs.init(); err != nil {
		cancel()
		return nil, fmt.Errorf("init chain broker service failed: %w", err)
	}

	if err := cbs.initWS(); err != nil {
		cancel()
		return nil, fmt.Errorf("init chain broker websocket service failed: %w", err)
	}

	return cbs, nil
}

func (cbs *ChainBrokerService) init() error {
	cbs.server = rpc.NewServer()

	apis, err := GetAPIs(cbs.rep, cbs.api, cbs.logger)
	if err != nil {
		return fmt.Errorf("get apis failed: %w", err)
	}

	// Register all the APIs exposed by the namespace services
	for _, api := range apis {
		if err := cbs.server.RegisterName(api.Namespace, api.Service); err != nil {
			return fmt.Errorf("register name %s for service %v failed: %w", api.Namespace, api.Service, err)
		}
	}

	return nil
}

func (cbs *ChainBrokerService) initWS() error {
	cbs.wsServer = rpc.NewServer()

	apis, err := GetAPIs(cbs.rep, cbs.api, cbs.logger)
	if err != nil {
		return fmt.Errorf("get apis failed: %w", err)
	}

	// Register all the APIs exposed by the namespace services
	for _, api := range apis {
		if err := cbs.wsServer.RegisterName(api.Namespace, api.Service); err != nil {
			return fmt.Errorf("register name %s for service %v failed: %w", api.Namespace, api.Service, err)
		}
	}

	return nil
}

func (cbs *ChainBrokerService) Start() error {
	router := mux.NewRouter()
	handler := cbs.tokenBucketMiddleware(cbs.server)
	router.Handle("/", handler)

	wsRouter := mux.NewRouter()
	wsHandler := node.NewWSHandlerStack(cbs.wsServer.WebsocketHandler([]string{"*"}), []byte(""))
	wsRouter.Handle("/", wsHandler)

	go func() {
		cbs.logger.WithFields(logrus.Fields{
			"port": cbs.rep.Config.Port.JsonRpc,
		}).Info("JSON-RPC service started")

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cbs.rep.Config.Port.JsonRpc), cors.Default().Handler(router)); err != nil {
			cbs.logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Errorf("Failed to start JSON_RPC service: %s", err.Error())
			return
		}
	}()

	go func() {
		cbs.logger.WithFields(logrus.Fields{
			"port": cbs.rep.Config.Port.WebSocket,
		}).Info("Websocket service started")

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cbs.rep.Config.Port.WebSocket), cors.Default().Handler(wsRouter)); err != nil {
			cbs.logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Errorf("Failed to start websocket service: %s", err.Error())
			return
		}
	}()

	return nil
}

func (cbs *ChainBrokerService) Stop() error {
	cbs.cancel()

	cbs.server.Stop()

	cbs.logger.Info("JSON-RPC service stopped")

	return nil
}

func (cbs *ChainBrokerService) tokenBucketMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readJsonRpc := cbs.rep.Config.JsonRPC.ReadLimiter.Enable
		writeJsonRpc := cbs.rep.Config.JsonRPC.WriteLimiter.Enable

		if !readJsonRpc && !writeJsonRpc {
			next.ServeHTTP(w, r)
		} else {
			requestBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Restore the r.Body with the captured content.
			r.Body = io.NopCloser(bytes.NewReader(requestBody))

			newRequest, err := http.NewRequest(r.Method, r.URL.String(), io.NopCloser(bytes.NewReader(requestBody)))
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			newRequest.Header = r.Header.Clone()

			var request any
			var isJson = true
			if err := json.NewDecoder(io.NopCloser(bytes.NewReader(requestBody))).Decode(&request); err != nil {
				isJson = false
				cbs.logger.Error("tokenBucketMiddleware JSON decode error: ", err)
			}

			var rateLimiter *ratelimiter.JRateLimiter
			if isJson {
				switch req := request.(type) {
				case []any:
					for _, req := range req {
						if reqMap, ok := req.(map[string]any); ok {
							method, ok := reqMap["method"].(string)
							if ok {
								cbs.logger.Info("request method: ", method)
								if method == "eth_sendRawTransaction" {
									if writeJsonRpc {
										rateLimiter = cbs.rateLimiterForWrite
									}
									cbs.logger.Info("tokenBucketMiddleware rateLimiter is rateLimiterForWrite writeJsonRpc:", writeJsonRpc)
								} else if readJsonRpc {
									rateLimiter = cbs.rateLimiterForRead
								}
							}
						}
					}
				default:
					if readJsonRpc {
						rateLimiter = cbs.rateLimiterForRead
					}
				}
			}

			if rateLimiter != nil && rateLimiter.JLimit() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, newRequest)
		}
	})
}
