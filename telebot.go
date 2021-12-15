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

	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

var Bot *tb.Bot
var TOKEN = ""

var GROUPS = []int64{}
var ADMINS = []int64{}

var lastID = int64(-1)
var lastText = ""
var puncReg *regexp.Regexp

var zcomap *ObliviousMap

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

func IsGroupAdmin(m *tb.Message) bool {
	gc := GetGroupConfig(m.Chat.ID)
	return gc != nil && gc.IsAdmin(m.Sender.ID)
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

	// ---------------- Super Admin ----------------

	Bot.Handle("/su_add_group", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
			if UpdateGroup(m.Chat.ID, UMAdd) {
				SmartSend(m, "✔️ 已将该组加入积分统计 ～")
			} else {
				SmartSend(m, "❌ 该组已经开启积分统计啦 ～")
			}
		}
	})

	Bot.Handle("/su_del_group", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
			if UpdateGroup(m.Chat.ID, UMDel) {
				SmartSend(m, "✔️ 已将该组移除积分统计 ～")
			} else {
				SmartSend(m, "❌ 该组尚未开启积分统计哦 ～")
			}
		}
	})

	Bot.Handle("/su_add_admin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
				SmartSend(m.ReplyTo, "✔️ TA 已经成为管理员啦 ～")
			} else {
				SmartSend(m.ReplyTo, "❌ TA 已经是管理员啦 ～")
			}
		}
	})

	Bot.Handle("/su_del_admin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSend(m.ReplyTo, "✔️ 已将 TA 的管理员移除 ～")
			} else {
				SmartSend(m.ReplyTo, "❌ TA 本来就不是管理员呢 ～")
			}
		}
	})

	// ---------------- Group Admin ----------------

	Bot.Handle("/addadmin", func(m *tb.Message) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
			if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
				SmartSend(m.ReplyTo, "✔️ TA 已经成为群管理员啦 ～")
			} else {
				SmartSend(m.ReplyTo, "❌ TA 已经是群管理员啦 ～")
			}
		}
	})

	Bot.Handle("/deladmin", func(m *tb.Message) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
			if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSend(m.ReplyTo, "✔️ 已将 TA 的群管理员移除 ～")
			} else {
				SmartSend(m.ReplyTo, "❌ TA 本来就不是群管理员呢 ～")
			}
		}
	})

	Bot.Handle("/setcredit", func(m *tb.Message) {
		if IsGroupAdmin(m) {
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
				target = BuildCreditInfo(m.Chat.ID, m.ReplyTo.Sender, false)
			}
			target = UpdateCredit(target, UMSet, credit)
			SmartSend(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.Credit))
		} else {
			SmartSend(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
	})

	Bot.Handle("/addcredit", func(m *tb.Message) {
		if IsGroupAdmin(m) {
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
				target = BuildCreditInfo(m.Chat.ID, m.ReplyTo.Sender, false)
			}
			target = UpdateCredit(target, UMAdd, credit)
			SmartSend(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.Credit))
		} else {
			SmartSend(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
	})

	Bot.Handle("/creditrank", func(m *tb.Message) {
		if IsGroupAdmin(m) {
			rank, _ := strconv.Atoi(m.Payload)
			if rank <= 0 {
				rank = 10
			} else if rank > 30 {
				rank = 30
			}
			ranks := GetCreditRank(m.Chat.ID, rank)
			rankStr := ""
			for i, c := range ranks {
				rankStr += fmt.Sprintf("`%2d`. `%s`: `%d`\n", i+1, strings.ReplaceAll(c.Name, "`", "'"), c.Credit)
			}
			SmartSend(m, "👀 当前的积分墙为: \n\n"+rankStr)
		} else {
			SmartSend(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
	})

	Bot.Handle("/lottery", func(m *tb.Message) {
		if IsGroupAdmin(m) {
			rank, _ := strconv.Atoi(m.Payload)
			if rank <= 0 {
				rank = 10
			} else if rank > 100 {
				rank = 100
			}
			ranks := GetCreditRank(m.Chat.ID, rank)
			num := rand.Intn(len(ranks))
			c := ranks[num]
			rankStr := fmt.Sprintf(" [-](%s) `%s`\n", fmt.Sprintf("tg://user?id=%d", c.ID), strings.ReplaceAll(c.Name, "`", "'"))
			SmartSend(m, fmt.Sprintf("🎉 恭喜以下用户中奖：\n\n"+rankStr))
		} else {
			SmartSend(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
	})

	// ---------------- Normal User ----------------

	Bot.Handle("/ban", func(m *tb.Message) {
		if IsGroupAdmin(m) && ValidReplyUser(m) {
			if err := Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
				SmartSend(m, fmt.Sprintf("🎉 恭喜 %s 获得禁言大礼包，可喜可贺可喜可贺！", GetUserName(m.ReplyTo.Sender)))
			} else {
				DErrorE(err, "Perm Update | Fail to ban user")
				SmartSend(m, "❌ 您没有办法禁言 TA 呢")
			}
		}
	})

	Bot.Handle("/unban", func(m *tb.Message) {
		if IsGroupAdmin(m) && ValidReplyUser(m) {
			if err := Unban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
				SmartSend(m, fmt.Sprintf("🎉 恭喜 %s 重新获得了自由 ～", GetUserName(m.ReplyTo.Sender)))
			} else {
				DErrorE(err, "Perm Update | Fail to unban user")
				SmartSend(m, "❌ 您没有办法解禁 TA 呢")
			}
		}
	})

	Bot.Handle("/mycredit", func(m *tb.Message) {
		if m.Chat.ID > 0 {
			SmartSend(m, "❌ 请在群组发送这条命令来查看积分哦 ～")
		} else {
			SmartSend(m, fmt.Sprintf("👀 您当前的积分为: %d", GetCredit(m.Chat.ID, m.Sender.ID).Credit))
		}
	})

	Bot.Handle(tb.OnUserLeft, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.UserLeft.ID > 0 && !m.UserLeft.IsBot {
				UpdateCredit(BuildCreditInfo(m.Chat.ID, m.UserLeft, false), UMSet, 0)
			}
		}
	})

	Bot.Handle("嘴臭", func(m *tb.Message) {
		if IsGroup(m.Chat.ID) && m.ReplyTo != nil {
			if m.Sender.ID > 0 && m.Sender.Username != "Channel_Bot" {
				if m.ReplyTo.Sender.ID == m.Sender.ID {
					SmartSend(m, "确实")
				} else if m.ReplyTo.Sender.ID < 0 || m.ReplyTo.Sender.IsBot {
					SmartSend(m, "它没嘴呢 ...")
				} else {
					token := fmt.Sprintf("%d,%d,%d", m.Chat.ID, m.Sender.ID, m.ReplyTo.Sender.ID)
					if _, ok := zcomap.Get(token); ok {
						UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMAdd, -10)
						SmartSend(m, "😠 你自己先漱漱口呢，不要连续臭别人哦！扣 10 分警告一下")
					} else {
						zcomap.Set(token, 1)
						ci := UpdateCredit(BuildCreditInfo(m.Chat.ID, m.ReplyTo.Sender, false), UMAdd, -25)
						SmartSend(m.ReplyTo, fmt.Sprintf("您被 %s 警告了 ⚠️，请注意管理好自己的 Psycho-Pass！暂时扣除 25 分作为警告，如果您的分数低于 -50 分将被直接禁言。若您觉得这是恶意举报，请理性对待，并联系群管理员处理。", GetUserName(m.Sender)))
						Bot.Delete(m)
						if ci.Credit < -50 {
							Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0)
						}
					}
				}
			} else {
				SmartSend(m, "😠 匿名就不要乱啵啵啦！叭了个叭叭了个叭叭了个叭 ...")
			}
		}
	})

	Bot.Handle(tb.OnText, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.IsForwarded() {
				return
			}

			text := strings.TrimSpace(m.Text)
			textLen := len([]rune(text))
			userId := m.Sender.ID

			if m.Sender.Username != "Channel_Bot" {
				if puncReg.MatchString(text) {
					UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMAdd, -10)
					lastID = userId
				} else if textLen >= 2 {
					if lastID == userId && text == lastText {
						UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMAdd, -5)
					} else if lastID != userId || (textLen >= 14 && text != lastText) {
						UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMAdd, 1)
					}
					lastID = userId
					lastText = text
				}
			}

			if ValidReplyUser(m) {
				UpdateCredit(BuildCreditInfo(m.Chat.ID, m.ReplyTo.Sender, false), UMAdd, 1)
			}
		}
	})

	Bot.Handle(tb.OnSticker, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.IsForwarded() {
				return
			}
			userId := m.Sender.ID
			if m.Sender.Username != "Channel_Bot" {
				if lastID != userId {
					UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMAdd, 1)
					lastID = userId
				}
			}

			if ValidReplyUser(m) {
				UpdateCredit(BuildCreditInfo(m.Chat.ID, m.ReplyTo.Sender, false), UMAdd, 1)
			}
		}
	})

	go Bot.Start()
	DInfo("telegram bot is up.")
}

func ValidReplyUser(m *tb.Message) bool {
	return m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot &&
		m.ReplyTo.Sender.ID != m.Sender.ID && m.ReplyTo.Sender.Username != "Channel_Bot"
}

func BuildCreditInfo(groupId int64, user *tb.User, autoFetch bool) *CreditInfo {
	ci := &CreditInfo{
		user.Username, GetUserName(user), user.ID, 0, groupId,
	}
	if autoFetch {
		ci.Credit = GetCredit(groupId, user.ID).Credit
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
	s := ""
	if u.FirstName != "" || u.LastName != "" {
		s = strings.TrimSpace(u.FirstName + " " + u.LastName)
	} else if u.Username != "" {
		s = "@" + u.Username
	} else {
		s = fmt.Sprintf("%d", u.ID)
	}

	return s
}

func Ban(chatId, userId int64, duration int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		cm.CanSendMessages = false
		cm.CanSendMedia = false
		cm.CanSendOther = false
		cm.CanAddPreviews = false
		cm.CanSendPolls = false
		cm.CanInviteUsers = false
		cm.CanPinMessages = false
		cm.CanChangeInfo = false

		cm.RestrictedUntil = time.Now().Unix() + duration
		return RestrictChatMember(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func Unban(chatId, userId int64, duration int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		cm.CanSendMessages = true
		cm.CanSendMedia = true
		cm.CanSendOther = true
		cm.CanAddPreviews = true
		cm.CanSendPolls = true
		cm.CanInviteUsers = true
		cm.CanPinMessages = true
		cm.CanChangeInfo = true
		return RestrictChatMember(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func RestrictChatMember(chat *tb.Chat, member *tb.ChatMember) error {
	rights, until := member.Rights, member.RestrictedUntil

	params := map[string]interface{}{
		"chat_id":     chat.Recipient(),
		"user_id":     member.User.Recipient(),
		"permissions": &map[string]bool{},
		"until_date":  strconv.FormatInt(until, 10),
	}

	data, _ := jsoniter.Marshal(rights)
	_ = jsoniter.Unmarshal(data, params["permissions"])
	_, err := Bot.Raw("restrictChatMember", params)
	return err
}

func init() {
	puncReg = regexp.MustCompile(`^[!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`" + `][a-zA-Z0-9]+`)
	zcomap = NewOMap(60 * 60 * 1000)
}
