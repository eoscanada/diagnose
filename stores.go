package main

import (
	"github.com/eoscanada/bstream/store"
)

func (d *Diagnose) setupStores(blocksStoreURL, searchIndexesStoreURL string) error {
	searchStore, err := store.NewSimpleGStore(searchIndexesStoreURL)
	if err != nil {
		return err
	}

	blocksStore, err := store.NewSimpleArchiveStore(blocksStoreURL)
	if err != nil {
		return err
	}

	d.blocksStore = blocksStore
	d.searchStore = searchStore

	return nil
}
