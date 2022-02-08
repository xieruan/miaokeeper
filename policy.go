package main

import (
	"regexp"
	"sync"

	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

type GroupConfig struct {
	ID            int64
	Admins        []int64
	BannedForward []int64
	MergeTo       int64

	Locale           string
	MustFollow       string
	MustFollowOnJoin bool
	MustFollowOnMsg  bool

	AntiSpoiler bool
	DisableWarn bool

	WarnKeywords []string
	BanKeywords  []string

	NameBlackListReg   []string
	NameBlackListRegEx []*regexp.Regexp `json:"-"`

	updateLock sync.RWMutex `json:"-"`
}

func NewGroupConfig(groupId int64) *GroupConfig {
	return SetGroupConfig(groupId, (&GroupConfig{
		ID: groupId,
	}).Check())
}

func (gc *GroupConfig) Check() *GroupConfig {
	if gc == nil {
		gc = &GroupConfig{}
	}

	gc.updateLock.Lock()
	defer gc.updateLock.Unlock()

	if gc.Admins == nil {
		gc.Admins = make([]int64, 0)
	}
	if gc.BannedForward == nil {
		gc.BannedForward = make([]int64, 0)
	}
	if gc.WarnKeywords == nil {
		gc.WarnKeywords = make([]string, 0)
	}
	if gc.BanKeywords == nil {
		gc.BanKeywords = make([]string, 0)
	}
	if gc.NameBlackListReg == nil {
		gc.NameBlackListReg = make([]string, 0)
	}
	if gc.NameBlackListRegEx == nil {
		gc.NameBlackListRegEx = make([]*regexp.Regexp, 0)
		for _, regStr := range gc.NameBlackListReg {
			if regex, err := regexp.Compile(regStr); regex != nil && err == nil {
				gc.NameBlackListRegEx = append(gc.NameBlackListRegEx, regex)
			} else if err != nil {
				DErrorf("Name BlackList Error | Not compilable regex=%s err=%s", regStr, err.Error())
			}
		}
	}
	return gc
}

func (gc *GroupConfig) ToJson(indent bool) (s string) {
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	if !indent {
		s, _ = jsoniter.MarshalToString(gc)
	} else {
		if b, err := jsoniter.MarshalIndent(gc, "", "  "); err == nil && b != nil {
			s = string(b)
		}
	}
	return
}

func (gc *GroupConfig) FromJson(s string) error {
	gc.updateLock.Lock()
	defer gc.updateLock.Unlock()

	return jsoniter.UnmarshalFromString(s, gc)
}

func (gc *GroupConfig) Clone() *GroupConfig {
	newGC := GroupConfig{}
	newGC.FromJson(gc.ToJson(false))

	return (&newGC).Check()
}

func (gc *GroupConfig) UpdateAdmin(userId int64, method UpdateMethod) bool {
	gc.updateLock.Lock()
	defer gc.updateLock.Unlock()

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
	gc.updateLock.Lock()
	defer gc.updateLock.Unlock()

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
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	return I64In(&gc.Admins, userId)
}

func (gc *GroupConfig) IsBannedForward(id int64) bool {
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	return I64In(&gc.BannedForward, id)
}

func (gc *GroupConfig) IsBanKeyword(m *tb.Message) bool {
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	keywords := gc.BanKeywords
	if len(keywords) == 0 {
		keywords = DefaultBanKeywords
	}
	return ContainsString(keywords, m.Text)
}

func (gc *GroupConfig) IsWarnKeyword(m *tb.Message) bool {
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	keywords := gc.WarnKeywords
	if len(keywords) == 0 {
		keywords = DefaultWarnKeywords
	}
	return ContainsString(keywords, m.Text)
}

func (gc *GroupConfig) IsBlackListName(u *tb.User) bool {
	gc.updateLock.RLock()
	defer gc.updateLock.RUnlock()

	namePattern := u.LastName + u.FirstName + u.LastName + u.Username
	for _, regex := range gc.NameBlackListRegEx {
		if regex.MatchString(namePattern) {
			return true
		}
	}
	return false
}
