package ledger

import (
	"fmt"

	"github.com/axiomesh/axiom-ledger/internal/ledger/blockstm"
)

// Create a unique path for special fields (e.g. balance, code) in a state object.
// func subPath(prefix []byte, s uint8) [blockstm.KeyLength]byte {
// 	path := append(prefix, common.Hash{}.Bytes()...) // append a full empty hash to avoid collision with storage state
// 	path = append(path, s)                           // append the special field identifier

// 	return path
// }

const BalancePath = 1
const NoncePath = 2
const CodePath = 3
const SuicidePath = 4

func (s *StateLedgerImpl) SetMVHashmap(mvhm *blockstm.MVHashMap) {
	s.mvHashmap = mvhm
	s.dep = -1
}

func (s *StateLedgerImpl) GetMVHashmap() *blockstm.MVHashMap {
	return s.mvHashmap
}

func (s *StateLedgerImpl) MVWriteList() []blockstm.WriteDescriptor {
	writes := make([]blockstm.WriteDescriptor, 0, len(s.writeMap))

	for _, v := range s.writeMap {
		if _, ok := s.revertedKeys[v.Path]; !ok {
			writes = append(writes, v)
		}
	}

	return writes
}

func (s *StateLedgerImpl) MVFullWriteList() []blockstm.WriteDescriptor {
	writes := make([]blockstm.WriteDescriptor, 0, len(s.writeMap))

	for _, v := range s.writeMap {
		writes = append(writes, v)
	}

	return writes
}

func (s *StateLedgerImpl) MVReadMap() map[blockstm.Key]blockstm.ReadDescriptor {
	return s.readMap
}

func (s *StateLedgerImpl) MVReadList() []blockstm.ReadDescriptor {
	reads := make([]blockstm.ReadDescriptor, 0, len(s.readMap))

	for _, v := range s.MVReadMap() {
		reads = append(reads, v)
	}

	return reads
}

func (s *StateLedgerImpl) ensureReadMap() {
	if s.readMap == nil {
		s.readMap = make(map[blockstm.Key]blockstm.ReadDescriptor)
	}
}

func (s *StateLedgerImpl) ensureWriteMap() {
	if s.writeMap == nil {
		s.writeMap = make(map[blockstm.Key]blockstm.WriteDescriptor)
	}
}

func (s *StateLedgerImpl) ClearReadMap() {
	s.readMap = make(map[blockstm.Key]blockstm.ReadDescriptor)
}

func (s *StateLedgerImpl) ClearWriteMap() {
	s.writeMap = make(map[blockstm.Key]blockstm.WriteDescriptor)
}

func (s *StateLedgerImpl) HadInvalidRead() bool {
	return s.dep >= 0
}

func (s *StateLedgerImpl) DepTxIndex() int {
	return s.dep
}

func (s *StateLedgerImpl) SetIncarnation(inc int) {
	s.incarnation = inc
}

func MVRead[T any](s *StateLedgerImpl, k blockstm.Key, defaultV T, readStorage func(s *StateLedgerImpl) T) (v T) {
	if s.mvHashmap == nil {
		return readStorage(s)
	}

	s.ensureReadMap()

	if s.writeMap != nil {
		if _, ok := s.writeMap[k]; ok {
			return readStorage(s)
		}
	}

	if !k.IsAddress() {
		// If we are reading subpath from a deleted account, return default value instead of reading from MVHashmap
		addr := k.GetAddress()
		if s.GetAccount(addr) == nil {
			return defaultV
		}
	}

	res := s.mvHashmap.Read(k, s.txIndex)

	var rd blockstm.ReadDescriptor

	rd.V = blockstm.Version{
		TxnIndex:    res.DepIdx(),
		Incarnation: res.Incarnation(),
	}

	rd.Path = k

	switch res.Status() {
	case blockstm.MVReadResultDone:
		{
			v = readStorage(res.Value().(*StateLedgerImpl))
			rd.Kind = blockstm.ReadKindMap
		}
	case blockstm.MVReadResultDependency:
		{
			s.dep = res.DepIdx()

			panic("Found dependency")
		}
	case blockstm.MVReadResultNone:
		{
			v = readStorage(s)
			rd.Kind = blockstm.ReadKindStorage
		}
	default:
		return defaultV
	}

	// TODO: I assume we don't want to overwrite an existing read because this could - for example - change a storage
	//  read to map if the same value is read multiple times.
	if _, ok := s.readMap[k]; !ok {
		s.readMap[k] = rd
	}

	return
}

func MVRead2[T any, V any](s *StateLedgerImpl, k blockstm.Key, defaultV T, defaultV2 V, readStorage2 func(s *StateLedgerImpl) (T, V)) (v T, v2 V) {
	if s.mvHashmap == nil {
		return readStorage2(s)
	}

	s.ensureReadMap()

	if s.writeMap != nil {
		if _, ok := s.writeMap[k]; ok {
			return readStorage2(s)
		}
	}

	if !k.IsAddress() {
		// If we are reading subpath from a deleted account, return default value instead of reading from MVHashmap
		addr := k.GetAddress()
		if s.GetAccount(addr) == nil {
			return defaultV, defaultV2
		}
	}

	res := s.mvHashmap.Read(k, s.txIndex)

	var rd blockstm.ReadDescriptor

	rd.V = blockstm.Version{
		TxnIndex:    res.DepIdx(),
		Incarnation: res.Incarnation(),
	}

	rd.Path = k

	switch res.Status() {
	case blockstm.MVReadResultDone:
		{
			v, v2 = readStorage2(res.Value().(*StateLedgerImpl))
			rd.Kind = blockstm.ReadKindMap
		}
	case blockstm.MVReadResultDependency:
		{
			s.dep = res.DepIdx()

			panic("Found dependency")
		}
	case blockstm.MVReadResultNone:
		{
			v, v2 = readStorage2(s)
			rd.Kind = blockstm.ReadKindStorage
		}
	default:
		return defaultV, defaultV2
	}

	// TODO: I assume we don't want to overwrite an existing read because this could - for example - change a storage
	//  read to map if the same value is read multiple times.
	if _, ok := s.readMap[k]; !ok {
		s.readMap[k] = rd
	}

	return
}

func MVWrite(s *StateLedgerImpl, k blockstm.Key) {
	if s.mvHashmap != nil {
		s.ensureWriteMap()
		s.writeMap[k] = blockstm.WriteDescriptor{
			Path: k,
			V:    s.BlockStmVersion(),
			Val:  s,
		}
	}
}

func RevertWrite(s *StateLedgerImpl, k blockstm.Key) {
	s.revertedKeys[k] = struct{}{}
}

func MVWritten(s *StateLedgerImpl, k blockstm.Key) bool {
	if s.mvHashmap == nil || s.writeMap == nil {
		return false
	}

	_, ok := s.writeMap[k]

	return ok
}

// FlushMVWriteSet applies entries in the write set to MVHashMap. Note that this function does not clear the write set.
func (s *StateLedgerImpl) FlushMVWriteSet() {
	if s.mvHashmap != nil && s.writeMap != nil {
		s.mvHashmap.FlushMVWriteSet(s.MVFullWriteList())
	}
}

// ApplyMVWriteSet applies entries in a given write set to StateDB. Note that this function does not change MVHashMap nor write set
// of the current StateDB.
func (s *StateLedgerImpl) ApplyMVWriteSet(writes []blockstm.WriteDescriptor) {
	for i := range writes {
		path := writes[i].Path
		sr := writes[i].Val.(*StateLedgerImpl)

		if path.IsState() {
			addr := path.GetAddress()
			stateKey := path.GetStateKey()
			_, state := sr.GetState(addr, stateKey.Bytes())
			s.SetState(addr, stateKey.Bytes(), state)
		} else if path.IsAddress() {
			continue
		} else {
			addr := path.GetAddress()

			switch path.GetSubpath() {
			case BalancePath:
				s.SetBalance(addr, sr.GetBalance(addr))
			case NoncePath:
				s.SetNonce(addr, sr.GetNonce(addr))
			case CodePath:
				s.SetCode(addr, sr.GetCode(addr))
			case SuicidePath:
				stateObject := sr.GetAccount(addr)
				if stateObject != nil && stateObject.SelfDestructed() {
					s.SelfDestruct(addr)
				}
			default:
				panic(fmt.Errorf("unknown key type: %d", path.GetSubpath()))
			}
		}
	}
}

type DumpStruct struct {
	TxIdx  int
	TxInc  int
	VerIdx int
	VerInc int
	Path   []byte
	Op     string
}

// GetReadMapDump gets readMap Dump of format: "TxIdx, Inc, Path, Read"
func (s *StateLedgerImpl) GetReadMapDump() []DumpStruct {
	readList := s.MVReadList()
	res := make([]DumpStruct, 0, len(readList))

	for _, val := range readList {
		temp := &DumpStruct{
			TxIdx:  s.txIndex,
			TxInc:  s.incarnation,
			VerIdx: val.V.TxnIndex,
			VerInc: val.V.Incarnation,
			Path:   val.Path[:],
			Op:     "Read\n",
		}
		res = append(res, *temp)
	}

	return res
}

// GetWriteMapDump gets writeMap Dump of format: "TxIdx, Inc, Path, Write"
func (s *StateLedgerImpl) GetWriteMapDump() []DumpStruct {
	writeList := s.MVReadList()
	res := make([]DumpStruct, 0, len(writeList))

	for _, val := range writeList {
		temp := &DumpStruct{
			TxIdx:  s.txIndex,
			TxInc:  s.incarnation,
			VerIdx: val.V.TxnIndex,
			VerInc: val.V.Incarnation,
			Path:   val.Path[:],
			Op:     "Write\n",
		}
		res = append(res, *temp)
	}

	return res
}

// AddEmptyMVHashMap adds empty MVHashMap to StateDB
func (s *StateLedgerImpl) AddEmptyMVHashMap() {
	mvh := blockstm.MakeMVHashMap()
	s.mvHashmap = mvh
}

func (s *StateLedgerImpl) BlockStmVersion() blockstm.Version {
	return blockstm.Version{
		TxnIndex:    s.txIndex,
		Incarnation: s.incarnation,
	}
}
