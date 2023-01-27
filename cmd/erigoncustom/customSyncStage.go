package main

import (
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/eth"
	"github.com/ledgerwatch/erigon/eth/stagedsync"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	"github.com/ledgerwatch/log/v3"
)

var CustomSyncStage stages.SyncStage = "CustomSyncStage"

func CustomStage(chainDB kv.RwDB) *stagedsync.Stage {
	return &stagedsync.Stage{
		ID:          CustomSyncStage,
		Description: "Custom stage",
		Forward: func(firstCycle bool, badBlockUnwind bool, s *stagedsync.StageState, u stagedsync.Unwinder, tx kv.RwTx, quiet bool) (err error) {
			useExternalTx := tx != nil
			if !useExternalTx {
				tx, err = chainDB.BeginRw(context.Background())
				if err != nil {
					return err
				}
				defer tx.Rollback()
			}

			prevStageProgress, errStart := stages.GetStageProgress(tx, stages.Senders)
			if errStart != nil {
				return errStart
			}
			if s.BlockNumber+1 >= prevStageProgress {
				return nil
			}

			log.Info(fmt.Sprintf("[%s] Progress", s.LogPrefix()), "fromBlockNumber", s.BlockNumber+1, "toBlockNumber", prevStageProgress)

			if err := s.Update(tx, prevStageProgress); err != nil {
				return err
			}

			if !useExternalTx {
				if err = tx.Commit(); err != nil {
					return err
				}
			}
			return nil
		},
		Unwind: func(firstCycle bool, u *stagedsync.UnwindState, s *stagedsync.StageState, tx kv.RwTx) error {
			return nil
		},
		Prune: func(firstCycle bool, u *stagedsync.PruneState, tx kv.RwTx) error {
			return nil
		},
	}
}

func insertCustomSyncStage(ethereum *eth.Ethereum) {
	if ethereum.StagedSync() != nil {
		panic("Can't change stages in initiated Ethereum node")
	}

	syncUnwindOrder := ethereum.SyncUnwindOrder()
	if len(syncUnwindOrder) != 14 ||
		syncUnwindOrder[8] != stages.Translation ||
		syncUnwindOrder[9] != stages.Execution {
		panic("unknown syncUnwindOrder")
	}
	syncUnwindOrder = append(syncUnwindOrder[:8+1], syncUnwindOrder[8:]...)
	syncUnwindOrder[9] = CustomSyncStage

	syncPruneOrder := ethereum.SyncPruneOrder()
	if len(syncPruneOrder) != 15 ||
		syncPruneOrder[9] != stages.Translation ||
		syncPruneOrder[10] != stages.Execution {
		panic("unknown syncPruneOrder")
	}
	syncPruneOrder = append(syncPruneOrder[:9+1], syncPruneOrder[9:]...)
	syncPruneOrder[10] = CustomSyncStage

	syncStages := ethereum.SyncStages()
	if len(syncStages) != 16 ||
		syncStages[6].ID != stages.Execution ||
		syncStages[7].ID != stages.HashState {
		panic("unknown syncStages")
	}
	syncStages = append(syncStages[:6+1], syncStages[6:]...)
	syncStages[7] = CustomStage(ethereum.ChainDB())

	ethereum.SetSyncUnwindOrder(syncUnwindOrder)
	ethereum.SetSyncPruneOrder(syncPruneOrder)
	ethereum.SetSyncStages(syncStages)
}
