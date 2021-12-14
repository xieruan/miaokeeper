package main

import (
	"database/sql"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DBCONN = ""
var MYSQLDB *sql.DB
var UpdateLock sync.Mutex

type UpdateMethod int

const (
	UMSet UpdateMethod = iota
	UMAdd
	UMDel
)

type CreditInfo struct {
	Username     string
	Name         string
	ID           int64
	GlobalCredit int64
}

func InitDatabase() (err error) {
	MYSQLDB, err = sql.Open("mysql", DBCONN)
	if MYSQLDB != nil {
		MYSQLDB.SetConnMaxLifetime(time.Minute * 3)
		MYSQLDB.SetMaxOpenConns(20)
	}

	if err == nil {
		err = MYSQLDB.Ping()
	}

	return
}

func InitTables() {
	_, err := MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Config (
		k VARCHAR(128) NOT NULL PRIMARY KEY,
		v TEXT NOT NULL
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Table Creation Error | error=%v", err.Error())
		os.Exit(1)
	}
}

func ReadConfig(key string) string {
	ret := ""
	err := MYSQLDB.QueryRow(`SELECT v FROM MiaoKeeper_Config WHERE k = ?;`, key).Scan(&ret)
	if err != nil {
		DErrorf("Config Read Error | error=%v", err.Error())
	}
	return ret
}

func WriteConfig(key, value string) {
	_, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Config
			(k, v)
		VALUES
			(?, ?)
		ON DUPLICATE KEY UPDATE
			v = VALUES(v)`, key, value)
	if err != nil {
		DErrorf("Config Write Error | error=%v", err.Error())
	}
}

func InitCreditTable() {
	_, err := MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Credit_Global (
		userid BIGINT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		username TEXT NOT NULL,
		credit BIGINT NOT NULL
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Table Creation Error | error=%v", err.Error())
	}
}

func ReadConfigs() {
	ADMINS = ParseStrToInt64Arr(ReadConfig("ADMINS"))
	GROUPS = ParseStrToInt64Arr(ReadConfig("GROUPS"))
}

func WriteConfigs() {
	WriteConfig("ADMINS", ParseInt64ArrToStr(ADMINS))
	WriteConfig("GROUPS", ParseInt64ArrToStr(GROUPS))
}

func UpdateAdmin(userId int64, method UpdateMethod) {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	if method == UMSet {
		ADMINS = []int64{userId}
	} else if method == UMAdd {
		ADMINS = AddIntoInt64Arr(ADMINS, userId)
	} else if method == UMDel {
		ADMINS = DelFromInt64Arr(ADMINS, userId)
	}
	WriteConfigs()
}

func UpdateGroup(groupId int64, method UpdateMethod) {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	if method == UMSet {
		GROUPS = []int64{groupId}
	} else if method == UMAdd {
		GROUPS = AddIntoInt64Arr(GROUPS, groupId)
	} else if method == UMDel {
		GROUPS = DelFromInt64Arr(GROUPS, groupId)
	}
	WriteConfigs()
}

func GetCredit(userId int64) *CreditInfo {
	ret := &CreditInfo{}
	err := MYSQLDB.QueryRow(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_Global WHERE userid = ?;`, userId).Scan(
		&ret.ID, &ret.Name, &ret.Username, &ret.GlobalCredit,
	)
	if err != nil {
		DErrorE(err, "Database Credit Read Error")
	}
	return ret
}

func GetCreditRank(limit int) []*CreditInfo {
	returns := []*CreditInfo{}
	row, _ := MYSQLDB.Query(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_Global ORDER BY credit DESC LIMIT ?;`, limit)
	for row.Next() {
		ret := &CreditInfo{}
		row.Scan(&ret.ID, &ret.Name, &ret.Username, &ret.GlobalCredit)
		returns = append(returns, ret)
	}
	row.Close()
	return returns
}

func UpdateCredit(user *CreditInfo, method UpdateMethod, value int64) *CreditInfo {
	ci := GetCredit(user.ID)
	if user.Name == "" {
		user.Name = ci.Name
	}
	if user.Username == "" {
		user.Username = ci.Username
	}
	user.GlobalCredit = ci.GlobalCredit
	if method == UMSet {
		user.GlobalCredit = value
	} else if method == UMAdd {
		user.GlobalCredit += value
	} else if method == UMDel {
		user.GlobalCredit = 0
	}

	_, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Credit_Global
			(userid, name, username, credit)
		VALUES
			(?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			username = VALUES(username),
			credit = VALUES(credit)
		`, user.ID, user.Name, user.Username, user.GlobalCredit)
	if err != nil {
		DErrorE(err, "Database Credit Update Error")
	}

	return user
}
