package main

import (
	"github.com/eoscanada/dstore"
)

func (d *Diagnose) setupStores(blocksStoreURL, searchIndexesStoreURL string) error {
	searchStore, err := dstore.NewSimpleStore(searchIndexesStoreURL)
	if err != nil {
		return err
	}

	blocksStore, err := dstore.NewJSONLStore(blocksStoreURL)
	if err != nil {
		return err
	}

	d.blocksStore = blocksStore
	d.searchStore = searchStore

	return nil
}
