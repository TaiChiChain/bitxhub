package ledger

import (
	"bytes"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/axiomesh/axiom-kit/types"
)

type AccountCache struct {
	innerAccountCache     *lru.Cache[string, *InnerAccount]
	codeCache             *lru.Cache[string, []byte]
	enableExpensiveMetric bool
	disable               bool

	cacheMissCountPerBlock int
	cacheHitCountPerBlock  int
}

func NewAccountCache(cacheSize int, disable bool) (*AccountCache, error) {
	if disable {
		return &AccountCache{
			disable: true,
		}, nil
	}

	if cacheSize == 0 {
		cacheSize = 1024
	}
	innerAccountCache, err := lru.New[string, *InnerAccount](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("init innerAccountCache failed: %w", err)
	}

	codeCache, err := lru.New[string, []byte](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("init codeCache failed: %w", err)
	}

	return &AccountCache{
		innerAccountCache: innerAccountCache,
		codeCache:         codeCache,
		disable:           false,
	}, nil
}

func (ac *AccountCache) SetEnableExpensiveMetric(enable bool) {
	ac.enableExpensiveMetric = enable
}

func (ac *AccountCache) resetMetrics() {
	if ac.disable {
		return
	}

	ac.cacheHitCountPerBlock = 0
	ac.cacheMissCountPerBlock = 0
}

func (ac *AccountCache) exportMetrics() {
	if ac.disable {
		return
	}

	accountCacheHitCounterPerBlock.Set(float64(ac.cacheHitCountPerBlock))
	accountCacheMissCounterPerBlock.Set(float64(ac.cacheMissCountPerBlock))
}

func (ac *AccountCache) add(accounts map[string]IAccount) {
	if ac.disable {
		return
	}

	for addr, acc := range accounts {
		account := acc.(*SimpleAccount)

		if account.dirtyAccount != nil {
			ac.innerAccountCache.Add(addr, account.dirtyAccount)
		}

		if !bytes.Equal(account.originCode, account.dirtyCode) {
			ac.codeCache.Add(addr, account.dirtyCode)
		}
	}
}

func (ac *AccountCache) getInnerAccount(addr *types.Address) (*InnerAccount, bool) {
	if ac.disable {
		return nil, false
	}

	ret, ok := ac.innerAccountCache.Get(addr.String())
	if ok {
		ac.cacheHitCountPerBlock++
	} else {
		ac.cacheMissCountPerBlock++
	}
	return ret, ok
}

func (ac *AccountCache) getCode(addr *types.Address) ([]byte, bool) {
	if ac.disable {
		return nil, false
	}

	ret, ok := ac.codeCache.Get(addr.String())
	if ok {
		ac.cacheHitCountPerBlock++
	} else {
		ac.cacheMissCountPerBlock++
	}
	return ret, ok
}

func (ac *AccountCache) clear() {
	if ac.disable {
		return
	}

	ac.innerAccountCache.Purge()
	ac.codeCache.Purge()
}
