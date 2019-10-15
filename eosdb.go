package main

import (
	"github.com/eoscanada/kvdb"
	"github.com/eoscanada/kvdb/eosdb"
)

func (d *Diagnose) setupEOSDB(connectionInfo string) error {
	info, err := kvdb.NewConnectionInfo(connectionInfo)
	if err != nil {
		return err
	}

	db, err := eosdb.New(info.TablePrefix, info.Project, info.Instance, false)
	if err != nil {
		return err
	}

	d.eosdb = db

	return nil
}
