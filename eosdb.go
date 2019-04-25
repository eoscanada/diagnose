package main

import (
	"github.com/eoscanada/eosdb/bigtable"
)

func (d *Diagnose) setupEOSDB(connectionInfo string) error {
	info, err := bigtable.NewConnectionInfo(connectionInfo)
	if err != nil {
		return err
	}

	bigtable, err := bigtable.New(info.TablePrefix, info.Project, info.Instance, false)
	if err != nil {
		return err
	}

	d.bigtable = bigtable
	d.eosdb = bigtable

	return nil
}
