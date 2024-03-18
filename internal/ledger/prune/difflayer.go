package prune

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

type diffLayer struct {
	height uint64

	cache map[string]types.Node

	// todo implement here in a more elegant way
	accountCache map[string]types.Node
	storageCache map[string]types.Node

	ledgerStorage storage.Storage
}

const TypeAccount = 1
const TypeStorage = 2

func (dl *diffLayer) GetFromTrieCache(nodeKey []byte) (types.Node, bool) {
	if v, ok := dl.cache[string(nodeKey)]; ok {
		return v, true
	}
	return nil, false
}

func (dl *diffLayer) String() string {
	res := strings.Builder{}
	res.WriteString("Version[")
	res.WriteString(strconv.Itoa(int(dl.height)))
	res.WriteString(fmt.Sprintf("], \nJournals[\n"))
	res.WriteString(fmt.Sprintf("journal=%v\n", dl.cache))
	res.WriteString("]")
	return res.String()
}

// TODO: implement count journal size function
