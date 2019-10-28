package main

import (
	"github.com/eoscanada/kvdb"
	"github.com/eoscanada/kvdb/eosdb"
	"github.com/eoscanada/kvdb/ethdb"
)

func (d *Diagnose) SetupDB(connectionInfo string, protocol string) error {
	info, err := kvdb.NewConnectionInfo(connectionInfo)
	if err != nil {
		return err
	}

	switch protocol {
	case "EOS":
		db, err := eosdb.New(info.TablePrefix, info.Project, info.Instance, false)
		if err != nil {
			return err
		}
		d.EOSdb = db
	case "ETH":
		db, err := ethdb.New(info.TablePrefix, info.Project, info.Instance, false)
		if err != nil {
			return err
		}
		d.ETHdb = db
	}
	return nil
}
