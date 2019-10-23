package main

import (
	"github.com/eoscanada/kvdb"
	"github.com/eoscanada/kvdb/ethdb"
)

func (d *Diagnose) setupETHDB(connectionInfo string) error {
	info, err := kvdb.NewConnectionInfo(connectionInfo)
	if err != nil {
		return err
	}

	db, err := ethdb.New(info.TablePrefix, info.Project, info.Instance, false)
	if err != nil {
		return err
	}

	d.ethdb = db

	return nil
}
