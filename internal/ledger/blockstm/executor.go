package blockstm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/sirupsen/logrus"
)

type ExecResult struct {
	Err      error
	Ver      Version
	TxIn     TxnInput
	TxOut    TxnOutput
	TxAllOut TxnOutput
}

type Status string

const (
	// StatusPending tasks are ready for execution
	// all executing tasks are in pending state
	StatusPending Status = "pending"
	// StatusExecuted tasks are ready for validation
	// these tasks did not abort during execution
	StatusExecuted Status = "executed"
	// StatusAborted means the task has been aborted
	// these tasks transition to pending upon next execution
	StatusAborted Status = "aborted"
	// StatusValidated means the task has been validated
	// tasks in this status can be reset if an earlier task fails validation
	StatusValidated Status = "validated"
	// StatusWaiting tasks are waiting for another tx to complete
	// StatusWaiting Status = "waiting"
)

type ExecTask interface {
	Execute(mvh *MVHashMap) error
	MVReadList() []ReadDescriptor
	MVWriteList() []WriteDescriptor
	MVFullWriteList() []WriteDescriptor
	Hash() types.Hash
	Sender() types.Address
	Settle()
	Dependencies() []int
	GetIndex() int
	GetExecuteRes() ExecResult
	IsStatus(s Status) bool
	GetStatus() Status
	SetStatus(s Status)
	Increment()
}

type ErrExecAbortError struct {
	Dependency  int
	OriginError error
}

func (e ErrExecAbortError) Error() string {
	if e.Dependency >= 0 {
		return fmt.Sprintf("Execution aborted due to dependency %d", e.Dependency)
	} else {
		return "Execution aborted"
	}
}

type ParallelExecutor struct {
	tasks []ExecTask

	executeCh  chan func()
	validateCh chan func()

	// Number of workers that execute transactions speculatively
	workers int

	// An integer that tracks the index of last settled transaction
	lastSettled int

	// Multi-version hash map
	mvh *MVHashMap

	// Stores the inputs and outputs of the last incardanotion of all transactions
	lastTxIO *TxnInputOutput

	log logrus.FieldLogger
}

func NewParallelExecutor(tasks []ExecTask, workers int) *ParallelExecutor {
	numTasks := len(tasks)
	if workers < 1 {
		workers = numTasks
	}
	pe := &ParallelExecutor{
		tasks:       tasks,
		lastSettled: -1,
		workers:     workers,
		mvh:         MakeMVHashMap(),
		lastTxIO:    MakeTxnInputOutput(numTasks),
		log:         loggers.Logger(loggers.Executor),
	}

	return pe
}

type PropertyCheck func(*ParallelExecutor) error

func ExecuteParallel(tasks []ExecTask, workers int, interruptCtx context.Context) error {
	if len(tasks) == 0 {
		return nil
	}

	pe := NewParallelExecutor(tasks, workers)
	pe.executeCh = make(chan func(), len(tasks))
	pe.validateCh = make(chan func(), len(tasks))

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// execution tasks are limited by workers
	start(workerCtx, pe.executeCh, workers)

	// validation tasks uses length of tasks to avoid blocking on validation
	start(workerCtx, pe.validateCh, len(tasks))

	iterations := 0
	toExecute := tasks
	for !allValidated(tasks) {
		if err := pe.executeAll(iterations, toExecute); err != nil {
			return err
		}
		//Serial write execution result
		if err := pe.flushExecuteResToMvh(iterations, toExecute); err != nil {
			return err
		}
		var err error
		//Parallel Validation Tx
		toExecute, err = pe.validateAll(iterations, tasks)
		if err != nil {
			return err
		}
		startIdx, anyLeft := pe.findFirstNonValidated()
		if !anyLeft {
			pe.lastSettled = len(pe.tasks) - 1
		} else {
			pe.lastSettled = startIdx - 1
		}
		iterations++
	}
	pe.log.Info("len task", len(pe.tasks))
	pe.log.Info("iterations", iterations)
	start := time.Now()
	for _, task := range tasks {
		task.Settle()
	}
	blockStmSettleDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	return nil
}

func (pe *ParallelExecutor) executeAll(iterations int, tasks []ExecTask) error {
	start := time.Now()
	if len(tasks) == 0 {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for _, task := range tasks {
		t := task
		pe.DoExecute(func() {
			t.Execute(pe.mvh)
			t.SetStatus(StatusExecuted)
			wg.Done()
		})
	}
	wg.Wait()
	if iterations == 0 {
		blockStmExecuteDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	}
	return nil

}

func (pe *ParallelExecutor) DoExecute(work func()) {
	pe.executeCh <- work
}

func (pe *ParallelExecutor) DoValidate(work func()) {
	pe.validateCh <- work
}

func (pe *ParallelExecutor) flushExecuteResToMvh(iterations int, tasks []ExecTask) error {
	start := time.Now()
	for _, task := range tasks {
		res := task.GetExecuteRes()
		if res.Err == nil {
			pe.mvh.FlushMVWriteSet(res.TxAllOut)
			tx := res.Ver.TxnIndex
			pe.lastTxIO.recordRead(tx, res.TxIn)
			if res.Ver.Incarnation == 0 {
				pe.lastTxIO.recordWrite(tx, res.TxOut)
				pe.lastTxIO.recordAllWrite(tx, res.TxAllOut)
			} else {
				prevWrite := pe.lastTxIO.AllWriteSet(tx)
				cmpMap := make(map[Key]bool)
				for _, w := range res.TxAllOut {
					cmpMap[w.Path] = true
				}

				for _, v := range prevWrite {
					if _, ok := cmpMap[v.Path]; !ok {
						pe.mvh.Delete(v.Path, tx)
					}
				}
				pe.lastTxIO.recordWrite(tx, res.TxOut)
				pe.lastTxIO.recordAllWrite(tx, res.TxAllOut)
			}
		}
	}
	if iterations == 0 {
		blockStmFlushDataDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	}
	return nil

}

func (pe *ParallelExecutor) validateAll(iterations int, tasks []ExecTask) ([]ExecTask, error) {
	start := time.Now()
	var res []ExecTask
	var mx sync.Mutex

	startIdx, anyLeft := pe.findFirstNonValidated()
	if !anyLeft {
		return nil, nil
	}

	wg := &sync.WaitGroup{}
	for i := startIdx; i < len(tasks); i++ {
		task := tasks[i]
		wg.Add(1)
		pe.DoValidate(func() {
			if !pe.validateTask(task) {
				mx.Lock()
				defer mx.Unlock()
				task.Increment()
				task.SetStatus(StatusPending)
				res = append(res, task)
			}
			wg.Done()
		})
	}

	wg.Wait()
	if iterations == 0 {
		blockStmValidateDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	}
	return res, nil
}

func (pe *ParallelExecutor) validateTask(task ExecTask) bool {
	switch task.GetStatus() {
	case StatusAborted, StatusPending:
		return false
	case StatusExecuted, StatusValidated:
		txIndex := task.GetIndex()
		if txIndex == pe.lastSettled+1 {
			task.SetStatus(StatusValidated)
			return true
		}
		execRes := task.GetExecuteRes()
		if execRes.Err != nil {
			return false
		}
		if ValidateVersion(task.GetIndex(), pe.lastTxIO, pe.mvh) {
			task.SetStatus(StatusValidated)
			return true
		} else {
			for _, v := range pe.lastTxIO.AllWriteSet(txIndex) {
				pe.mvh.Delete(v.Path, txIndex)
				//pe.mvh.MarkEstimate(v.Path, txIndex)
			}
			return false
		}
		// case StatusWaiting:
		// 	return true

	}
	panic("unexpected status: " + task.GetStatus())
}

func allValidated(tasks []ExecTask) bool {
	for _, t := range tasks {
		if !t.IsStatus(StatusValidated) {
			return false
		}
	}
	return true
}

func (pe *ParallelExecutor) findFirstNonValidated() (int, bool) {
	for i, t := range pe.tasks {
		if !t.IsStatus(StatusValidated) {
			return i, true
		}
	}
	return 0, false
}

func start(ctx context.Context, ch chan func(), workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case work := <-ch:
					work()
				}
			}
		}()
	}
}
