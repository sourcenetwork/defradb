package db

// type Txn struct {
// 	rootstore ds.Datastore

// 	// wrapped DS
// 	datastore core.DSReaderWriter // wrapped /data namespace
// 	headstore core.DSReaderWriter // wrapped /heads namespace
// 	dagstore  core.DAGStore       // wrapped /blocks namespace
// }

// func (db *DB) newTxn(readonly bool) (Txn, error) {
// 	var txn Txn
// 	var err error

// 	// check if our datastore natively supports transactions or Batching
// 	txnStore, ok := db.rootstore.(ds.TxnDatastore)
// 	if ok { // we support transactions
// 		dstxn, err = txnStore.NewTransaction(readonly)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return txnToDsShim(txn), nil
// 	} else { // no txn

// 	}
// }
