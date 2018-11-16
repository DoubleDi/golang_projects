package db

import (
	"time"

	"../entity/stat"
)

func (db *DB) AddStat(stat *stat.Stat) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into stats values (?, ?, ?)", stat.UserID, stat.Action, time.Time(stat.TS))

	return err
}
