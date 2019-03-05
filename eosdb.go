package main

import (
	"github.com/eoscanada/eosdb/bigtable"
)

func (d *Diagnose) setupEOSDB(cnx string) error {
	info, err := bigtable.NewConnectionInfo(cnx)
	if err != nil {
		return err
	}

	zeDB, err := bigtable.New(info.TablePrefix, info.Project, info.Instance, false)
	if err != nil {
		return err
	}

	d.eosdb = zeDB

	return nil
}
