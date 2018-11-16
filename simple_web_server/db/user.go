package db

import (
	"log"
	"time"

	"../entity/user"
)

func (db *DB) AddUser(user *user.User) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into users values (?, ?, ?)", user.ID, user.Age, user.Sex)

	return err
}

func (db *DB) TopUsers(date1 time.Time, date2 time.Time, action string, limit int) ([]user.UsersByDate, error) {
	err := db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println(date1, date2)
	rows, err := db.Query(`
		select id, age, sex, date(ts) as date, count(*) as count from 
			(select id, age, sex from users join stats on (users.id=stats.user_id) 
			 where action = ? 
			 group by stats.user_id order by count(*) desc limit ?) as top_users 
		join stats on (top_users.id=stats.user_id)
		where date(stats.ts) >= ? and date(stats.ts) <= ?
		group by id, date(ts) order by date(ts) desc
	`, action, limit, date1, date2)
	defer rows.Close()

	topUsers := make([]user.UsersByDate, 0, limit)

	userDate := ""
	userRows := []user.User{}
	for rows.Next() {
		var (
			date        string
			currentUser user.User
		)
		err = rows.Scan(&currentUser.ID, &currentUser.Age, &currentUser.Sex, &date, &currentUser.Count)
		if err != nil {
			return nil, err
		}
		if userDate == "" {
			userDate = date
		}

		if date != userDate {
			topUsers = append(topUsers, user.UsersByDate{
				Date: userDate,
				Rows: userRows,
			})
			userDate = date
			userRows = []user.User{}
		}

		userRows = append(userRows, currentUser)
	}

	if userDate != "" {
		topUsers = append(topUsers, user.UsersByDate{
			Date: userDate,
			Rows: userRows,
		})
	}

	return topUsers, err
}
