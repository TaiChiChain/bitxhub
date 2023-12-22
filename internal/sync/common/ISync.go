package common

//go:generate mockgen -destination mock_sync/mock_sync.go -package mock_sync -source ISync.go
type Sync interface {
	Prepare(opts ...Option) (*PrepareData, error)

	SwitchMode(mode SyncMode) error
	Stop()

	// Commit in full mode, we only need to commit commitDataCache
	// in snapshot mode, we need to commit chain data(both commitDataCache and receipt)
	Commit() chan any

	StartSync(params *SyncParams, syncTaskDoneCh chan error) error
}

type ISyncConstructor interface {
	Mode() SyncMode

	Prepare(config *Config) (*PrepareData, error)

	PostCommitData(data []CommitData)

	Commit() chan any
}
