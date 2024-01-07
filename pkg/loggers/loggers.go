package loggers

import (
	"context"
	"fmt"
	"path/filepath"

	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	P2P            = "p2p"
	Consensus      = "consensus"
	Executor       = "executor"
	Governance     = "governance"
	App            = "app"
	API            = "api"
	CoreAPI        = "coreapi"
	Storage        = "storage"
	Profile        = "profile"
	Finance        = "finance"
	BlockSync      = "blocksync"
	Access         = "access"
	TxPool         = "txpool"
	Epoch          = "epoch"
	SystemContract = "system_contract"
)

var w = &LoggerWrapper{
	loggers: map[string]*logrus.Entry{
		P2P:            log.NewWithModule(P2P),
		Consensus:      log.NewWithModule(Consensus),
		Executor:       log.NewWithModule(Executor),
		Governance:     log.NewWithModule(Governance),
		App:            log.NewWithModule(App),
		API:            log.NewWithModule(API),
		CoreAPI:        log.NewWithModule(CoreAPI),
		Storage:        log.NewWithModule(Storage),
		Profile:        log.NewWithModule(Profile),
		Finance:        log.NewWithModule(Finance),
		Access:         log.NewWithModule(Access),
		TxPool:         log.NewWithModule(TxPool),
		BlockSync:      log.NewWithModule(BlockSync),
		Epoch:          log.NewWithModule(Epoch),
		SystemContract: log.NewWithModule(SystemContract),
	},
}

type LoggerWrapper struct {
	loggers map[string]*logrus.Entry
}

type ethHandler struct {
	logger *logrus.Entry
}

func (e *ethHandler) Log(r *ethlog.Record) error {
	fields := logrus.Fields{}
	fields["ctx"] = r.Ctx
	switch r.Lvl {
	case ethlog.LvlCrit:
		e.logger.WithFields(fields).Fatal(r.Msg)
	case ethlog.LvlError:
		e.logger.WithFields(fields).Error(r.Msg)
	case ethlog.LvlWarn:
		e.logger.WithFields(fields).Warning(r.Msg)
	case ethlog.LvlInfo:
		e.logger.WithFields(fields).Info(r.Msg)
	case ethlog.LvlDebug:
		e.logger.WithFields(fields).Debug(r.Msg)
	case ethlog.LvlTrace:
		e.logger.WithFields(fields).Trace(r.Msg)
	}
	return nil
}

func InitializeEthLog(logger *logrus.Entry) {
	ethlog.Root().SetHandler(&ethHandler{logger: logger})
}

func Initialize(ctx context.Context, rep *repo.Repo, persist bool) error {
	config := rep.Config
	err := log.Initialize(
		log.WithCtx(ctx),
		log.WithEnableCompress(config.Log.EnableCompress),
		log.WithReportCaller(config.Log.ReportCaller),
		log.WithEnableColor(config.Log.EnableColor),
		log.WithDisableTimestamp(config.Log.DisableTimestamp),
		log.WithPersist(persist),
		log.WithFilePath(filepath.Join(rep.RepoRoot, repo.LogsDirName)),
		log.WithFileName(config.Log.Filename),
		log.WithMaxAge(int(config.Log.MaxAge)),
		log.WithMaxSize(int(config.Log.MaxSize)),
		log.WithRotationTime(config.Log.RotationTime.ToDuration()),
	)
	if err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}

	m := make(map[string]*logrus.Entry)
	m[P2P] = log.NewWithModule(P2P)
	m[P2P].Logger.SetLevel(log.ParseLevel(config.Log.Module.P2P))
	m[Consensus] = log.NewWithModule(Consensus)
	m[Consensus].Logger.SetLevel(log.ParseLevel(config.Log.Module.Consensus))
	m[Executor] = log.NewWithModule(Executor)
	m[Executor].Logger.SetLevel(log.ParseLevel(config.Log.Module.Executor))
	m[Governance] = log.NewWithModule(Governance)
	m[Governance].Logger.SetLevel(log.ParseLevel(config.Log.Module.Governance))
	m[App] = log.NewWithModule(App)
	m[App].Logger.SetLevel(log.ParseLevel(config.Log.Level))
	m[API] = log.NewWithModule(API)
	m[API].Logger.SetLevel(log.ParseLevel(config.Log.Module.API))
	m[CoreAPI] = log.NewWithModule(CoreAPI)
	m[CoreAPI].Logger.SetLevel(log.ParseLevel(config.Log.Module.CoreAPI))
	m[Storage] = log.NewWithModule(Storage)
	m[Storage].Logger.SetLevel(log.ParseLevel(config.Log.Module.Storage))
	m[Profile] = log.NewWithModule(Profile)
	m[Profile].Logger.SetLevel(log.ParseLevel(config.Log.Module.Profile))
	m[Finance] = log.NewWithModule(Finance)
	m[Finance].Logger.SetLevel(log.ParseLevel(config.Log.Module.Finance))
	m[BlockSync] = log.NewWithModule(BlockSync)
	m[BlockSync].Logger.SetLevel(log.ParseLevel(config.Log.Module.BlockSync))
	m[Access] = log.NewWithModule(Access)
	m[Access].Logger.SetLevel(log.ParseLevel(config.Log.Module.Access))
	m[TxPool] = log.NewWithModule(TxPool)
	m[TxPool].Logger.SetLevel(log.ParseLevel(config.Log.Module.TxPool))
	m[Epoch] = log.NewWithModule(Epoch)
	m[Epoch].Logger.SetLevel(log.ParseLevel(config.Log.Module.Epoch))
	m[SystemContract] = log.NewWithModule(SystemContract)
	m[SystemContract].Logger.SetLevel(log.ParseLevel(config.Log.Module.SystemContract))

	w = &LoggerWrapper{loggers: m}
	InitializeEthLog(m[API])
	return nil
}

func Logger(name string) logrus.FieldLogger {
	return w.loggers[name]
}
