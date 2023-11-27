package adaptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestDBStore(t *testing.T) {
	ast := assert.New(t)

	storageTypes := []string{repo.ConsensusStorageTypeMinifile, repo.ConsensusStorageTypeRosedb}

	for _, storageType := range storageTypes {
		t.Run(storageType, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			adaptor := mockAdaptorWithStorageType(ctrl, t, storageType)

			err := adaptor.StoreState("test1", []byte("value1"))
			ast.Nil(err)
			err = adaptor.StoreState("test2", []byte("value2"))
			ast.Nil(err)
			err = adaptor.StoreState("test3", []byte("value3"))
			ast.Nil(err)

			_, err = adaptor.ReadStateSet("not found")
			ast.NotNil(err)

			ret, _ := adaptor.ReadStateSet("test")
			ast.Equal(3, len(ret))

			err = adaptor.DelState("test1")
			ast.Nil(err)
			_, err = adaptor.ReadState("test1")
			ast.NotNil(err, "not found")
			ret1, _ := adaptor.ReadState("test2")
			ast.Equal([]byte("value2"), ret1)

			_, err = adaptor.ReadEpochState("epoch")
			ast.NotNil(err)
			err = adaptor.StoreEpochState("epoch", []byte("epochinfo"))
			ast.Nil(err)
			epochInfo, err := adaptor.ReadEpochState("epoch")
			ast.Nil(err)
			ast.Equal([]byte("epochinfo"), epochInfo)
		})
	}
}
