package saccount

import (
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/ownable"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	ownerKey = "owner" // owner key
)

var _ ownable.Ownable = (*Ownable)(nil)

type Ownable struct {
	common.SystemContractBase

	owner *common.VMSlot[ethcommon.Address]
}

// nolint
func (o *Ownable) Init(owner ethcommon.Address) {
	if o.owner.Has() {
		return
	}
	o.owner.Put(owner)
}

func (o *Ownable) SetContext(ctx *common.VMContext) {
	o.SystemContractBase.SetContext(ctx)

	o.owner = common.NewVMSlot[ethcommon.Address](o.StateAccount, ownerKey)
}

// Owner implements ownable.Owner.
func (o *Ownable) Owner() (ethcommon.Address, error) {
	return o.owner.MustGet()
}

// TransferOwnership implements ownable.Ownable.
func (o *Ownable) TransferOwnership(newOwner ethcommon.Address) error {
	oldOwner, err := o.Owner()
	if err != nil {
		return err
	}

	if oldOwner != o.Ctx.From {
		return o.Revert(&ownable.ErrorOwnableUnauthorizedAccount{})
	}

	return o.owner.Put(newOwner)
}

func (o *Ownable) CheckOwner() bool {
	owner, err := o.Owner()
	if err != nil {
		return false
	}

	if owner != o.Ctx.From {
		return false
	}

	return true
}
