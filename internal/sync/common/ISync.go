package common

//go:generate mockgen -destination mock_sync/mock_sync.go -package mock_sync -source ISync.go
type Sync interface {
	Prepare(opts ...Option) (*PrepareData, error)

	Start()
	SwitchMode(mode SyncMode) error
	Stop()

	// Commit in full mode, we only need to commit commitData
	// in snapshot mode, we need to commit chain data(both commitData and receipt)
	Commit() chan any

	StartSync(params *SyncParams, syncTaskDoneCh chan error) error

	GetSyncProgress() *SyncProgress
}

type ISyncConstructor interface {
	Mode() SyncMode

	Start()

	Prepare(config *Config) (*PrepareData, error)

	PostCommitData(data []CommitData)

	Commit() chan any

	Stop()
}
