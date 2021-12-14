package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var Bot *tb.Bot
var TOKEN = ""

var GROUPS = []int64{}
var ADMINS = []int64{}

var lastID = int64(-1)
var lastText = ""
var puncReg *regexp.Regexp

func SetCommands() error {
	allCommands := [][]string{
		{"mycredit", "获取自己的积分"},
		{"creditrank", "获取积分排行榜前 N 名"},
		{"lottery", "在积分排行榜前 N 名内抽出一名幸运儿"},
	}
	cmds := []tb.Command{}
	for _, cmd := range allCommands {
		cmds = append(cmds, tb.Command{
			Text:        cmd[0],
			Description: cmd[1],
		})
	}
	return Bot.SetCommands(cmds)
}

func IsGroup(gid int64) bool {
	return I64In(&GROUPS, gid)
}

func IsAdmin(uid int64) bool {
	return I64In(&ADMINS, uid)
}

func InitTelegram() {
	var err error
	Bot, err = tb.NewBot(tb.Settings{
		Token:  TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		DErrorE(err, "TeleBot Error | cannot initialize telegram bot")
		os.Exit(1)
	}

	err = SetCommands()
	if err != nil {
		DErrorE(err, "TeleBot Error | cannot update commands for telegram bot")
	}

	Bot.Handle("/mycredit", func(m *tb.Message) {
		if m.Chat.ID < 0 {
			SmartSend(m, "❌ 请私聊我来查看积分哦")
		} else {
			SmartSend(m, fmt.Sprintf("👀 您当前的积分为: %d", GetCredit(m.Sender.ID).GlobalCredit))
		}
	})

	Bot.Handle("/addgroup", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
			UpdateGroup(m.Chat.ID, UMAdd)
			SmartSend(m, "✔️ 已将该组加入积分统计 ～")
		}
	})

	Bot.Handle("/delgroup", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
			UpdateGroup(m.Chat.ID, UMDel)
			SmartSend(m, "✔️ 已将该组移除积分统计 ～")
		}
	})

	Bot.Handle("/addadmin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd)
			SmartSend(m.ReplyTo, "✔️ TA 已经成为管理员啦 ～")
		}
	})

	Bot.Handle("/deladmin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			UpdateAdmin(m.ReplyTo.Sender.ID, UMDel)
			SmartSend(m.ReplyTo, "✔️ 已将 TA 的管理员移除 ～")
		}
	})

	Bot.Handle("/setcredit", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) {
			addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
			target := &CreditInfo{}
			credit := int64(0)

			if len(addons) == 0 {
				SmartSend(m, "❌ 使用方法错误：/setcredit <UserId:Optional> <Credit>")
				return
			}

			if len(addons) == 1 {
				credit = addons[0]
			} else {
				target.ID = addons[0]
				credit = addons[1]
			}

			if m.ReplyTo != nil {
				target = BuildCreditInfo(m.ReplyTo.Sender, false)
			}
			target = UpdateCredit(target, UMSet, credit)
			SmartSend(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.GlobalCredit))
		}
	})

	Bot.Handle("/addcredit", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) {
			addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
			target := &CreditInfo{}
			credit := int64(0)

			if len(addons) == 0 {
				SmartSend(m, "❌ 使用方法错误：/addcredit <UserId:Optional> <Credit>")
				return
			}

			if len(addons) == 1 {
				credit = addons[0]
			} else {
				target.ID = addons[0]
				credit = addons[1]
			}

			if m.ReplyTo != nil {
				target = BuildCreditInfo(m.ReplyTo.Sender, false)
			}
			target = UpdateCredit(target, UMAdd, credit)
			SmartSend(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.GlobalCredit))
		}
	})

	Bot.Handle("/creditrank", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) {
			rank, _ := strconv.Atoi(m.Payload)
			if rank <= 0 {
				rank = 10
			} else if rank > 30 {
				rank = 30
			}
			ranks := GetCreditRank(rank)
			rankStr := ""
			for i, c := range ranks {
				rankStr += fmt.Sprintf("`%2d`. `%s`: `%d`\n", i+1, strings.ReplaceAll(c.Name, "`", "'"), c.GlobalCredit)
			}
			SmartSend(m, "👀 当前的积分墙为: \n\n"+rankStr)
		} else {
			SmartSend(m, "❌ 您没有查看积分墙的权限哦")
		}
	})

	Bot.Handle("/lottery", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) {
			rank, _ := strconv.Atoi(m.Payload)
			if rank <= 0 {
				rank = 10
			} else if rank > 100 {
				rank = 100
			}
			ranks := GetCreditRank(rank)
			num := rand.Intn(len(ranks))
			c := ranks[num]
			rankStr := fmt.Sprintf(" [-](%s) `%s`\n", fmt.Sprintf("tg://user?id=%d", c.ID), strings.ReplaceAll(c.Name, "`", "'"))
			SmartSend(m, fmt.Sprintf("🎉 恭喜以下用户中奖：\n\n"+rankStr))
		} else {
			SmartSend(m, "❌ 您没有查看积分墙的权限哦")
		}
	})

	// Bot.Handle(tb.OnUserLeft, func(m *tb.Message) {
	// 	if IsGroup(m.Chat.ID) {
	// 		userId := m.Sender.ID
	// 		SetCredit(userId, GetUserName(m.Sender), 0)
	// 	}
	// })

	Bot.Handle(tb.OnText, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.IsForwarded() {
				return
			}

			text := strings.TrimSpace(m.Text)
			textLen := len([]rune(text))
			userId := m.Sender.ID

			if puncReg.MatchString(text) {
				UpdateCredit(BuildCreditInfo(m.Sender, false), UMAdd, -10)
				lastID = userId
			} else if textLen >= 2 {
				if lastID == userId && text == lastText {
					UpdateCredit(BuildCreditInfo(m.Sender, false), UMAdd, -5)
				} else if lastID != userId || (textLen >= 14 && text != lastText) {
					UpdateCredit(BuildCreditInfo(m.Sender, false), UMAdd, 1)
				}
				lastID = userId
				lastText = text
			}

			if m.ReplyTo != nil && m.ReplyTo.Sender.ID != userId {
				UpdateCredit(BuildCreditInfo(m.ReplyTo.Sender, false), UMAdd, 2)
			}
		}
	})

	Bot.Handle(tb.OnSticker, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.IsForwarded() {
				return
			}
			userId := m.Sender.ID
			if lastID != userId {
				UpdateCredit(BuildCreditInfo(m.Sender, false), UMAdd, 1)
				lastID = userId
			}

			if m.ReplyTo != nil && m.ReplyTo.Sender.ID != userId {
				UpdateCredit(BuildCreditInfo(m.ReplyTo.Sender, false), UMAdd, 1)
			}
		}
	})

	go Bot.Start()
	DInfo("telegram bot is up.")
}

func BuildCreditInfo(user *tb.User, autoFetch bool) *CreditInfo {
	ci := &CreditInfo{
		user.Username, GetUserName(user), user.ID, 0,
	}
	if autoFetch {
		ci.GlobalCredit = GetCredit(user.ID).GlobalCredit
	}
	return ci
}

func SmartSend(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	if len(options) == 0 {
		return SmartSendInner(to, what, &tb.SendOptions{
			ParseMode:             "Markdown",
			DisableWebPagePreview: true,
		})
	}
	return SmartSendInner(to, what, options...)
}

func SmartSendInner(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	toType := Type(to)
	var m *tb.Message = nil
	var err error = nil
	if toType == "*telebot.Message" {
		mess, _ := to.(*tb.Message)
		m, err = Bot.Reply(mess, what, options...)
	} else if toType == "*telebot.Chat" {
		recp, _ := to.(*tb.Chat)
		if recp != nil {
			m, err = Bot.Send(recp, what, options...)
		} else {
			err = errors.New("chat is empty")
		}
	} else if toType == "*telebot.User" {
		recp, _ := to.(*tb.User)
		if recp != nil {
			m, err = Bot.Send(recp, what, options...)
		} else {
			err = errors.New("user is empty")
		}
	} else if toType == "int64" {
		recp, _ := to.(int64)
		m, err = Bot.Send(&tb.Chat{ID: recp}, what, options...)
	} else {
		err = errors.New("unknown type of message: " + toType)
	}
	if err != nil {
		DErrorE(err, "TeleBot Message Error")
	}
	return m, err
}

func GetUserName(u *tb.User) string {
	if u.FirstName != "" || u.LastName != "" {
		return strings.TrimSpace(u.FirstName + " " + u.LastName)
	} else if u.Username != "" {
		return "@" + u.Username
	} else {
		return fmt.Sprintf("%d", u.ID)
	}
}

func init() {
	puncReg = regexp.MustCompile(`^[!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`]")
}
