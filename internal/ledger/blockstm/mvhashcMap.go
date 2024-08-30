package blockstm

import (
	"fmt"
	"sync"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/emirpasic/gods/maps/treemap"

	"github.com/ethereum/go-ethereum/common"

	cmap "github.com/orcaman/concurrent-map/v2"
)

const FlagDone = 0
const FlagEstimate = 1

const addressType = 1
const stateType = 2
const subpathType = 3

const KeyLength = common.AddressLength + common.HashLength + 2

type Key [KeyLength]byte

func (k Key) IsAddress() bool {
	return k[KeyLength-1] == addressType
}

func (k Key) IsState() bool {
	return k[KeyLength-1] == stateType
}

func (k Key) IsSubpath() bool {
	return k[KeyLength-1] == subpathType
}

func (k Key) String() string {
	return string(k[:])
}

func (k Key) GetAddress() *types.Address {
	return types.NewAddress(common.BytesToAddress(k[:common.AddressLength]).Bytes())
}

func (k Key) GetStateKey() *types.Hash {
	return types.NewHash(common.BytesToHash(k[common.AddressLength : KeyLength-2]).Bytes())
}

func (k Key) GetKeyType() byte {
	return k[KeyLength-1]
}

func (k Key) GetSubpath() byte {
	return k[KeyLength-2]
}

func newKey(addr types.Address, hash types.Hash, subpath byte, keyType byte) Key {
	var k Key

	copy(k[:common.AddressLength], addr.Bytes())
	copy(k[common.AddressLength:KeyLength-2], hash.Bytes())
	k[KeyLength-2] = subpath
	k[KeyLength-1] = keyType

	return k
}

func NewAddressKey(addr *types.Address) Key {
	if addr == nil {
		return newKey(types.Address{}, types.Hash{}, 0, addressType)
	}
	return newKey(*addr, types.Hash{}, 0, addressType)
}

func NewStateKey(addr *types.Address, hash types.Hash) Key {
	var k Key
	if addr == nil {
		k = newKey(types.Address{}, hash, 0, stateType)
	} else {
		k = newKey(*addr, hash, 0, stateType)
	}
	if !k.IsState() {
		panic(fmt.Errorf("key is not a state key"))
	}

	return k
}

func NewSubpathKey(addr *types.Address, subpath byte) Key {
	if addr == nil {
		return newKey(types.Address{}, types.Hash{}, subpath, subpathType)
	}
	return newKey(*addr, types.Hash{}, subpath, subpathType)
}

type MVHashMap struct {
	lock sync.Mutex
	m    cmap.ConcurrentMap[string, *TxnIndexCells]
}

func MakeMVHashMap() *MVHashMap {
	return &MVHashMap{
		m: cmap.New[*TxnIndexCells](),
	}
}

type WriteCell struct {
	flag        uint
	incarnation int
	data        interface{}
}

type TxnIndexCells struct {
	rw sync.RWMutex
	tm *treemap.Map
}

type Version struct {
	TxnIndex    int
	Incarnation int
}

func (mv *MVHashMap) getKeyCells(k Key, fNoKey func(kenc Key) *TxnIndexCells) (cells *TxnIndexCells) {
	// val, ok := mv.m.Load(k)
	val, ok := mv.m.Get(k.String())
	if !ok {
		cells = fNoKey(k)
	} else {
		cells = val
	}

	return
}

func (mv *MVHashMap) Write(k Key, v Version, data interface{}) {
	// start := time.Now()
	// defer func() {
	// 	mvHashmapWriteDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	// }()
	cells := mv.getKeyCells(k, func(kenc Key) (cells *TxnIndexCells) {
		mv.lock.Lock()
		if val, ok := mv.m.Get(k.String()); !ok {
			n := &TxnIndexCells{
				rw: sync.RWMutex{},
				tm: treemap.NewWithIntComparator(),
			}
			mv.m.Set(k.String(), n)
			cells = n
		} else {
			cells = val
		}

		mv.lock.Unlock()
		return
	})

	cells.rw.Lock()
	if ci, ok := cells.tm.Get(v.TxnIndex); !ok {
		cells.tm.Put(v.TxnIndex, &WriteCell{
			flag:        FlagDone,
			incarnation: v.Incarnation,
			data:        data,
		})
	} else {
		if ci.(*WriteCell).incarnation > v.Incarnation {
			panic(fmt.Errorf("existing transaction value does not have lower incarnation: %v, %v",
				k, v.TxnIndex))
		}
		ci.(*WriteCell).flag = FlagDone
		ci.(*WriteCell).incarnation = v.Incarnation
		ci.(*WriteCell).data = data
	}
	cells.rw.Unlock()
}

func (mv *MVHashMap) MarkEstimate(k Key, txIdx int) {
	cells := mv.getKeyCells(k, func(_ Key) *TxnIndexCells {
		panic(fmt.Errorf("path must already exist"))
	})

	cells.rw.Lock()
	if ci, ok := cells.tm.Get(txIdx); !ok {
		panic(fmt.Sprintf("should not happen - cell should be present for path. TxIdx: %v, path, %x, cells keys: %v", txIdx, k, cells.tm.Keys()))
	} else {
		ci.(*WriteCell).flag = FlagEstimate
	}
	cells.rw.Unlock()
}

func (mv *MVHashMap) Delete(k Key, txIdx int) {
	cells := mv.getKeyCells(k, func(_ Key) *TxnIndexCells {
		panic(fmt.Errorf("path must already exist"))
	})

	cells.rw.Lock()
	defer cells.rw.Unlock()
	cells.tm.Remove(txIdx)
}

const (
	MVReadResultDone       = 0
	MVReadResultDependency = 1
	MVReadResultNone       = 2
)

type MVReadResult struct {
	depIdx      int
	incarnation int
	value       interface{}
}

func (res *MVReadResult) DepIdx() int {
	return res.depIdx
}

func (res *MVReadResult) Incarnation() int {
	return res.incarnation
}

func (res *MVReadResult) Value() interface{} {
	return res.value
}

func (mvr MVReadResult) Status() int {
	if mvr.depIdx != -1 {
		if mvr.incarnation == -1 {
			return MVReadResultDependency
		} else {
			return MVReadResultDone
		}
	}

	return MVReadResultNone
}

func (mv *MVHashMap) Read(k Key, txIdx int) (res MVReadResult) {
	// start := time.Now()
	// defer func() {
	// 	mvHashmapReadDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	// }()

	res.depIdx = -1
	res.incarnation = -1

	cells := mv.getKeyCells(k, func(_ Key) *TxnIndexCells {
		return nil
	})
	if cells == nil {
		return
	}

	cells.rw.RLock()

	fk, fv := cells.tm.Floor(txIdx - 1)

	if fk != nil && fv != nil {
		c := fv.(*WriteCell)
		switch c.flag {
		case FlagEstimate:
			res.depIdx = fk.(int)
			res.value = c.data
		case FlagDone:
			{
				res.depIdx = fk.(int)
				res.incarnation = c.incarnation
				res.value = c.data
			}
		default:
			panic(fmt.Errorf("should not happen - unknown flag value"))
		}
	}

	cells.rw.RUnlock()

	return
}

func (mv *MVHashMap) FlushMVWriteSet(writes []WriteDescriptor) {
	for _, v := range writes {
		mv.Write(v.Path, v.V, v.Val)
	}
}

func ValidateVersion(txIdx int, lastInputOutput *TxnInputOutput, versionedData *MVHashMap) (valid bool) {
	valid = true

	for _, rd := range lastInputOutput.ReadSet(txIdx) {
		mvResult := versionedData.Read(rd.Path, txIdx)
		switch mvResult.Status() {
		case MVReadResultDone:
			valid = rd.Kind == ReadKindMap && rd.V == Version{
				TxnIndex:    mvResult.depIdx,
				Incarnation: mvResult.incarnation,
			}
		case MVReadResultDependency:
			valid = false
		case MVReadResultNone:
			valid = rd.Kind == ReadKindStorage // feels like an assertion?
		default:
			panic(fmt.Errorf("should not happen - undefined mv read status: %ver", mvResult.Status()))
		}

		if !valid {
			break
		}
	}

	return
}
