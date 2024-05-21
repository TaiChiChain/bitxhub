package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
)

type Address struct {
	Street string
	City   string
}

type Person struct {
	Name         string
	Age          uint
	HomeAddress  Address
	PhoneNumbers []Address
	Attributes   map[string]Address
}

type Company struct {
	Name      string
	CEO       Person
	Employees []Person
}

func TestGenerateSolidityStruct(t *testing.T) {
	solidityCode, err := GenerateSolidityStruct(Company{})
	assert.Nil(t, err)
	t.Log(solidityCode)

	solidityCode, err = GenerateSolidityStruct(&types.EpochInfo{})
	assert.Nil(t, err)
	t.Log(solidityCode)
}
