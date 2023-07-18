package main

import (
	"database/sql"
	"fmt"
	"log"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
)

type CommitExecuter struct {
	Twopc_info *pb.TwoPhaseCommitInfo
}

func (ex *CommitExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to commit prepared transacion Id: %s}\n",
		ex.Twopc_info.TransactionId)

	preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.Twopc_info.TransactionId, rdb.shardNum, rdb.replicaNum)
	commitTxRaw := fmt.Sprintf("COMMIT PREPARED '%s'", preparedTxId)

	commitStmt := RawStatement(commitTxRaw)

	_, err = commitStmt.Exec(db)

	if err != nil {
		log.Printf("failed to commit prepared tx {Id: %s}.\n", ex.Twopc_info.TransactionId)
		return false, err
	}

	log.Printf("successfully commited prepared tx {Id: %s}\n", ex.Twopc_info.TransactionId)
	return true, nil
}



type RollbackExecuter struct {
	Twopc_info *pb.TwoPhaseCommitInfo
}

func (ex *RollbackExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to rollback prepared transacion Id: %s}\n",
		ex.Twopc_info.TransactionId)

	preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.Twopc_info.TransactionId, rdb.shardNum, rdb.replicaNum)
	rollbackTxRaw := fmt.Sprintf("ROLLBACK PREPARED '%s'", preparedTxId)

	commitStmt := RawStatement(rollbackTxRaw)

	_, err = commitStmt.Exec(db)

	if err != nil {
		log.Printf("failed to rollback prepared tx {Id: %s}.\n", ex.Twopc_info.TransactionId)
		return false, err
	}
	log.Printf("successfully rolledback prepared tx {Id: %s}\n", ex.Twopc_info.TransactionId)
	return true, nil
}
