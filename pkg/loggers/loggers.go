package loggers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	"github.com/sirupsen/logrus"

	kitlog "github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	P2P            = "p2p"
	Consensus      = "consensus"
	Executor       = "executor"
	App            = "app"
	API            = "api"
	CoreAPI        = "coreapi"
	Storage        = "storage"
	Profile        = "profile"
	Finance        = "finance"
	BlockSync      = "blocksync"
	TxPool         = "txpool"
	SystemContract = "system_contract"
)

var w = &LoggerWrapper{
	loggers: map[string]*logrus.Entry{
		P2P:            kitlog.NewWithModule(P2P),
		Consensus:      kitlog.NewWithModule(Consensus),
		Executor:       kitlog.NewWithModule(Executor),
		App:            kitlog.NewWithModule(App),
		API:            kitlog.NewWithModule(API),
		CoreAPI:        kitlog.NewWithModule(CoreAPI),
		Storage:        kitlog.NewWithModule(Storage),
		Profile:        kitlog.NewWithModule(Profile),
		Finance:        kitlog.NewWithModule(Finance),
		TxPool:         kitlog.NewWithModule(TxPool),
		BlockSync:      kitlog.NewWithModule(BlockSync),
		SystemContract: kitlog.NewWithModule(SystemContract),
	},
}

type LoggerWrapper struct {
	loggers map[string]*logrus.Entry
}

func InitializeEthLog(logger *logrus.Entry) {
	log.SetDefault(log.NewLogger(&LogrusHandler{
		Logger: logger,
		Level:  levelMapReverse[logger.Level],
	}))
}

func Initialize(ctx context.Context, rep *repo.Repo, persist bool) error {
	config := rep.Config
	err := kitlog.Initialize(
		kitlog.WithCtx(ctx),
		kitlog.WithEnableCompress(config.Log.EnableCompress),
		kitlog.WithReportCaller(config.Log.ReportCaller),
		kitlog.WithEnableColor(config.Log.EnableColor),
		kitlog.WithDisableTimestamp(config.Log.DisableTimestamp),
		kitlog.WithPersist(persist),
		kitlog.WithFilePath(filepath.Join(rep.RepoRoot, repo.LogsDirName)),
		kitlog.WithFileName(config.Log.Filename),
		kitlog.WithMaxAge(int(config.Log.MaxAge)),
		kitlog.WithMaxSize(int(config.Log.MaxSize)),
		kitlog.WithRotationTime(config.Log.RotationTime.ToDuration()),
	)
	if err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}

	m := make(map[string]*logrus.Entry)
	m[P2P] = kitlog.NewWithModule(P2P)
	m[P2P].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.P2P))
	m[Consensus] = kitlog.NewWithModule(Consensus)
	m[Consensus].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.Consensus))
	m[Executor] = kitlog.NewWithModule(Executor)
	m[Executor].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.Executor))
	m[App] = kitlog.NewWithModule(App)
	m[App].Logger.SetLevel(kitlog.ParseLevel(config.Log.Level))
	m[API] = kitlog.NewWithModule(API)
	m[API].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.API))
	m[CoreAPI] = kitlog.NewWithModule(CoreAPI)
	m[CoreAPI].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.CoreAPI))
	m[Storage] = kitlog.NewWithModule(Storage)
	m[Storage].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.Storage))
	m[Profile] = kitlog.NewWithModule(Profile)
	m[Profile].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.Profile))
	m[Finance] = kitlog.NewWithModule(Finance)
	m[Finance].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.Finance))
	m[BlockSync] = kitlog.NewWithModule(BlockSync)
	m[BlockSync].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.BlockSync))
	m[TxPool] = kitlog.NewWithModule(TxPool)
	m[TxPool].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.TxPool))
	m[SystemContract] = kitlog.NewWithModule(SystemContract)
	m[SystemContract].Logger.SetLevel(kitlog.ParseLevel(config.Log.Module.SystemContract))

	w = &LoggerWrapper{loggers: m}
	InitializeEthLog(m[API])
	return nil
}

func Logger(name string) logrus.FieldLogger {
	return w.loggers[name]
}
