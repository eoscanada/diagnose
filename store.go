package main

import (
	"github.com/eoscanada/dstore"
	"go.uber.org/zap"
)

func (d *Diagnose) SetupStores(blocksStoreURL, searchIndexesStoreURL string) error {
	searchStore, err := dstore.NewSimpleStore(searchIndexesStoreURL)
	if err != nil {
		zlog.Fatal("failed setting up search store", zap.Error(err))
	}

	blocksStore, err := dstore.NewDBinStore(blocksStoreURL)
	if err != nil {
		zlog.Fatal("failed setting up block store", zap.Error(err))
	}

	d.BlocksStore = blocksStore
	d.SearchStore = searchStore

	return nil
}
