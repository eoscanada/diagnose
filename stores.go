package main

import (
	"github.com/eoscanada/dstore"
)

func (d *Diagnose) setupStores(blocksStoreURL, searchIndexesStoreURL string) error {
	searchStore, err := dstore.NewSimpleStore(searchIndexesStoreURL)
	if err != nil {
		return err
	}

	blocksStore, err := dstore.NewDBinStore(blocksStoreURL)
	if err != nil {
		return err
	}

	d.BlocksStore = blocksStore
	d.SearchStore = searchStore

	return nil
}
