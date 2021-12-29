package main

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
)

var DBCONN = ""
var MYSQLDB *sql.DB
var UpdateLock sync.Mutex
var GroupConfigLock sync.Mutex

type UpdateMethod int

const (
	UMSet UpdateMethod = iota
	UMAdd
	UMDel
)

type CreditInfo struct {
	Username string
	Name     string
	ID       int64
	Credit   int64
	GroupId  int64
}

var GroupConfigCache map[int64]*GroupConfig

type GroupConfig struct {
	ID            int64
	Admins        []int64
	BannedForward []int64

	MustFollow       string
	MustFollowOnJoin bool
	MustFollowOnMsg  bool
}

func InitDatabase() (err error) {
	MYSQLDB, err = sql.Open("mysql", DBCONN)
	if err == nil {
		err = MYSQLDB.Ping()
	}

	return
}

func InitTables() {
	q, err := MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Config (
		k VARCHAR(128) NOT NULL PRIMARY KEY,
		v TEXT NOT NULL
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Table Creation Error | error=%v", err.Error())
		os.Exit(1)
	}
	if q != nil {
		q.Close()
	}
}

func ReadConfig(key string) string {
	ret := ""
	err := MYSQLDB.QueryRow(`SELECT v FROM MiaoKeeper_Config WHERE k = ?;`, key).Scan(&ret)
	if err != nil {
		DLogf("Config Read Error | key=%s error=%v", key, err.Error())
	}
	return ret
}

func WriteConfig(key, value string) {
	q, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Config
			(k, v)
		VALUES
			(?, ?)
		ON DUPLICATE KEY UPDATE
			v = VALUES(v)`, key, value)
	if err != nil {
		DLogf("Config Write Error | key=%s value=%s error=%v", key, value, err.Error())
	}
	if q != nil {
		q.Close()
	}
}

func GetGroupConfig(groupId int64) *GroupConfig {
	if !IsGroup(groupId) {
		return nil
	}

	GroupConfigLock.Lock()
	defer GroupConfigLock.Unlock()
	if gc, ok := GroupConfigCache[groupId]; ok {
		return gc
	}

	cfg := ReadConfig(fmt.Sprintf("gcfg_%d", Abs(groupId)))
	if cfg != "" {
		gc := &GroupConfig{}
		err := jsoniter.Unmarshal([]byte(cfg), gc)
		if err == nil {
			GroupConfigCache[groupId] = gc
			return gc
		}
	}
	return nil
}

func SetGroupConfig(groupId int64, gc *GroupConfig) *GroupConfig {
	if !IsGroup(groupId) {
		return nil
	}

	GroupConfigLock.Lock()
	defer GroupConfigLock.Unlock()
	GroupConfigCache[groupId] = gc

	cfg, _ := jsoniter.Marshal(gc)
	if len(cfg) > 0 {
		WriteConfig(fmt.Sprintf("gcfg_%d", Abs(groupId)), string(cfg))
	}
	return gc
}

func InitGroupTable(groupId int64) {
	q, err := MYSQLDB.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Credit_%d (
		userid BIGINT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		username TEXT NOT NULL,
		credit BIGINT NOT NULL
	) DEFAULT CHARSET=utf8mb4`, Abs(groupId)))
	if err != nil {
		DErrorf("Table Creation Error | error=%v", err.Error())
	}
	if q != nil {
		q.Close()
	}

	if GetGroupConfig(groupId) == nil {
		NewGroupConfig(groupId)
	}
}

func NewGroupConfig(groupId int64) *GroupConfig {
	return SetGroupConfig(groupId, &GroupConfig{
		ID:     groupId,
		Admins: make([]int64, 0),
		BannedForward: make([]int64, 0),
	})
}

func ReadConfigs() {
	ADMINS = ParseStrToInt64Arr(ReadConfig("ADMINS"))
	GROUPS = ParseStrToInt64Arr(ReadConfig("GROUPS"))
	for _, g := range GROUPS {
		InitGroupTable(g)
	}
}

func WriteConfigs() {
	WriteConfig("ADMINS", ParseInt64ArrToStr(ADMINS))
	WriteConfig("GROUPS", ParseInt64ArrToStr(GROUPS))
}

func UpdateAdmin(userId int64, method UpdateMethod) bool {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	changed := false
	if method == UMSet {
		if len(ADMINS) != 1 || ADMINS[0] != userId {
			changed = true
			ADMINS = []int64{userId}
		}
	} else if method == UMAdd {
		ADMINS, changed = AddIntoInt64Arr(ADMINS, userId)
	} else if method == UMDel {
		ADMINS, changed = DelFromInt64Arr(ADMINS, userId)
	}
	WriteConfigs()
	return changed
}

func (gc *GroupConfig) UpdateAdmin(userId int64, method UpdateMethod) bool {
	changed := false
	if method == UMSet {
		if len(gc.Admins) != 1 || gc.Admins[0] != userId {
			changed = true
			gc.Admins = []int64{userId}
		}
	} else if method == UMAdd {
		gc.Admins, changed = AddIntoInt64Arr(gc.Admins, userId)
	} else if method == UMDel {
		gc.Admins, changed = DelFromInt64Arr(gc.Admins, userId)
	}
	SetGroupConfig(gc.ID, gc)
	return changed
}

func (gc *GroupConfig) UpdateBannedForward(id int64, method UpdateMethod) bool {
	changed := false
	if method == UMSet {
		if len(gc.BannedForward) != 1 || gc.BannedForward[0] != id {
			changed = true
			gc.BannedForward = []int64{id}
		}
	} else if method == UMAdd {
		gc.BannedForward, changed = AddIntoInt64Arr(gc.BannedForward, id)
	} else if method == UMDel {
		gc.BannedForward, changed = DelFromInt64Arr(gc.BannedForward, id)
	}
	SetGroupConfig(gc.ID, gc)
	return changed
}

func (gc *GroupConfig) IsAdmin(userId int64) bool {
	return I64In(&gc.Admins, userId)
}

func (gc *GroupConfig) IsBannedForward(id int64) bool {
	return I64In(&gc.BannedForward, id)
}

func UpdateGroup(groupId int64, method UpdateMethod) bool {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	changed := false
	if method == UMSet {
		if len(GROUPS) != 1 || GROUPS[0] != groupId {
			changed = true
			GROUPS = []int64{groupId}
		}
	} else if method == UMAdd {
		GROUPS, changed = AddIntoInt64Arr(GROUPS, groupId)
	} else if method == UMDel {
		GROUPS, changed = DelFromInt64Arr(GROUPS, groupId)
	}
	if changed && method == UMSet || method == UMAdd {
		InitGroupTable(groupId)
	}
	WriteConfigs()
	return changed
}

func GetCredit(groupId, userId int64) *CreditInfo {
	ret := &CreditInfo{}
	err := MYSQLDB.QueryRow(fmt.Sprintf(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_%d WHERE userid = ?;`, Abs(groupId)), userId).Scan(
		&ret.ID, &ret.Name, &ret.Username, &ret.Credit,
	)
	if err != nil {
		DLogf("Database Credit Read Error | gid=%d uid=%d error=%s", groupId, userId, err.Error())
	}
	return ret
}

func GetCreditRank(groupId int64, limit int) []*CreditInfo {
	returns := []*CreditInfo{}
	row, _ := MYSQLDB.Query(fmt.Sprintf(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_%d ORDER BY credit DESC LIMIT ?;`, Abs(groupId)), limit)
	for row.Next() {
		ret := &CreditInfo{}
		row.Scan(&ret.ID, &ret.Name, &ret.Username, &ret.Credit)
		returns = append(returns, ret)
	}
	row.Close()
	return returns
}

func UpdateCredit(user *CreditInfo, method UpdateMethod, value int64) *CreditInfo {
	ci := GetCredit(user.GroupId, user.ID)
	if user.Name == "" {
		user.Name = ci.Name
	}
	if user.Username == "" {
		user.Username = ci.Username
	}
	user.Credit = ci.Credit
	if method == UMSet {
		user.Credit = value
	} else if method == UMAdd {
		user.Credit += value
	} else if method == UMDel {
		user.Credit = 0
	}

	var query *sql.Rows
	var err error

	if method != UMDel {
		query, err = MYSQLDB.Query(fmt.Sprintf(`INSERT INTO MiaoKeeper_Credit_%d
				(userid, name, username, credit)
			VALUES
				(?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				username = VALUES(username),
				credit = VALUES(credit)
			`, Abs(user.GroupId)), user.ID, user.Name, user.Username, user.Credit)
	} else {
		query, err = MYSQLDB.Query(fmt.Sprintf(`DELETE FROM MiaoKeeper_Credit_%d
			WHERE userid = ?;`, Abs(user.GroupId)), user.ID)
	}
	if err != nil {
		DErrorE(err, "Database Credit Update Error")
	}
	if query != nil {
		query.Close()
	}

	DLogf("Update Credit | group=%d user=%d alter=%d credit=%d", Abs(user.GroupId), user.ID, method, value)

	return user
}

func init() {
	GroupConfigCache = make(map[int64]*GroupConfig)
}
