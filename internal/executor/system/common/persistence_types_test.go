package common

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

func TestVMMap(t *testing.T) {
	type Value struct {
		Name string
		Desc string
	}

	account := ledger.NewMockAccount(1, types.NewAddressByStr(ZeroAddress))
	vmMap := NewVMMap[string, Value](account, "test", func(key string) string { return key })

	assert.False(t, vmMap.Has("test"))
	exist, v, err := vmMap.Get("test")
	assert.Nil(t, err)
	assert.Empty(t, v)
	assert.False(t, exist)

	old := Value{Name: "name", Desc: "desc"}
	err = vmMap.Put("test", old)
	assert.Nil(t, err)

	exist, v, err = vmMap.Get("test")
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v, old))
	assert.True(t, exist)

	newValue := Value{Name: "new name", Desc: "new desc"}
	err = vmMap.Put("test", newValue)
	assert.Nil(t, err)

	exist, v, err = vmMap.Get("test")
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v, newValue))
	assert.True(t, exist)
	assert.True(t, vmMap.Has("test"))

	err = vmMap.Put("nil_value", Value{})
	assert.Nil(t, err)
	assert.True(t, vmMap.Has("nil_value"))

	exist, v2, err := vmMap.Get("nil_value")
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v2, Value{}))
	assert.True(t, exist)

	err = vmMap.Delete("nil_value")
	assert.Nil(t, err)
	assert.False(t, vmMap.Has("nil_value"))
	exist, v2, err = vmMap.Get("nil_value")
	assert.Nil(t, err)
	assert.Empty(t, v2)
	assert.False(t, exist)

	err = vmMap.Put("nil_value", old)
	assert.Nil(t, err)
	assert.True(t, vmMap.Has("nil_value"))
	exist, v2, err = vmMap.Get("nil_value")
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v2, old))
	assert.True(t, exist)
}

func TestVMSlotStruct(t *testing.T) {
	type Value struct {
		Name string
		Desc string
	}

	account := ledger.NewMockAccount(1, types.NewAddressByStr(ZeroAddress))
	vmMap := NewVMSlot[Value](account, "test")

	assert.False(t, vmMap.Has())
	exist, v, err := vmMap.Get()
	assert.Nil(t, err)
	assert.Empty(t, v)
	assert.False(t, exist)

	old := Value{Name: "name", Desc: "desc"}
	err = vmMap.Put(old)
	assert.Nil(t, err)

	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v, old))
	assert.True(t, exist)

	newValue := Value{Name: "new name", Desc: "new desc"}
	err = vmMap.Put(newValue)
	assert.Nil(t, err)
	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v, newValue))
	assert.True(t, exist)

	err = vmMap.Delete()
	assert.Nil(t, err)
	assert.False(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.Empty(t, v)
	assert.False(t, exist)

	err = vmMap.Put(newValue)
	assert.Nil(t, err)
	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(v, newValue))
	assert.True(t, exist)
}

func TestVMSlotString(t *testing.T) {
	account := ledger.NewMockAccount(1, types.NewAddressByStr(ZeroAddress))
	vmMap := NewVMSlot[string](account, "test")

	assert.False(t, vmMap.Has())
	exist, v, err := vmMap.Get()
	assert.Nil(t, err)
	assert.Empty(t, v)
	assert.False(t, exist)

	err = vmMap.Put("value")
	assert.Nil(t, err)

	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.Equal(t, "value", v)
	assert.True(t, exist)

	err = vmMap.Put("value2")
	assert.Nil(t, err)
	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.Equal(t, "value2", v)
	assert.True(t, exist)

	err = vmMap.Delete()
	assert.Nil(t, err)
	assert.False(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.Empty(t, v)
	assert.False(t, exist)

	err = vmMap.Put("value2")
	assert.Nil(t, err)
	assert.True(t, vmMap.Has())
	exist, v, err = vmMap.Get()
	assert.Nil(t, err)
	assert.Equal(t, "value2", v)
	assert.True(t, exist)
}
