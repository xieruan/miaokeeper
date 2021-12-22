package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

type UIGStatus int

const (
	UIGIn UIGStatus = iota
	UIGOut
	UIGErr
)

var Bot *tb.Bot
var TOKEN = ""

var GROUPS = []int64{}
var ADMINS = []int64{}

var lastID = int64(-1)
var lastText = ""
var puncReg *regexp.Regexp

var zcomap *ObliviousMap
var creditomap *ObliviousMap
var votemap *ObliviousMap

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

func IsGroupAdmin(c *tb.Chat, u *tb.User) bool {
	isGAS := IsGroupAdminMiaoKo(c, u)
	if isGAS {
		return true
	}
	return IsGroupAdminTelegram(c, u)
}

func IsGroupAdminMiaoKo(c *tb.Chat, u *tb.User) bool {
	gc := GetGroupConfig(c.ID)
	return gc != nil && gc.IsAdmin(u.ID)
}

func IsGroupAdminTelegram(c *tb.Chat, u *tb.User) bool {
	cm, _ := Bot.ChatMemberOf(c, u)
	if cm != nil && (cm.Role == tb.Administrator || cm.Role == tb.Creator) {
		return true
	}
	return false
}

func LazyDelete(m *tb.Message) {
	time.AfterFunc(time.Second*10, func() {
		Bot.Delete(m)
	})
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
				SmartSendDelete(m, "✔️ 已将该组加入积分统计 ～")
			} else {
				SmartSendDelete(m, "❌ 该组已经开启积分统计啦 ～")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/su_del_group", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
			if UpdateGroup(m.Chat.ID, UMDel) {
				SmartSendDelete(m, "✔️ 已将该组移除积分统计 ～")
			} else {
				SmartSendDelete(m, "❌ 该组尚未开启积分统计哦 ～")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/su_add_admin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
				SmartSendDelete(m.ReplyTo, "✔️ TA 已经成为管理员啦 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 已经是管理员啦 ～")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/su_del_admin", func(m *tb.Message) {
		if IsAdmin(m.Sender.ID) && m.ReplyTo != nil && m.ReplyTo.Sender.ID > 0 && !m.ReplyTo.Sender.IsBot {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSendDelete(m.ReplyTo, "✔️ 已将 TA 的管理员移除 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 本来就不是管理员呢 ～")
			}
		}
		LazyDelete(m)
	})

	// ---------------- Group Admin ----------------

	Bot.Handle("/add_admin", func(m *tb.Message) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
			if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
				SmartSendDelete(m.ReplyTo, "✔️ TA 已经成为群管理员啦 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 已经是群管理员啦 ～")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/del_admin", func(m *tb.Message) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
			if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSendDelete(m.ReplyTo, "✔️ 已将 TA 的群管理员移除 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 本来就不是群管理员呢 ～")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/set_credit", func(m *tb.Message) {
		if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
			addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
			target := &CreditInfo{}
			credit := int64(0)

			if len(addons) == 0 {
				SmartSendDelete(m, "❌ 使用方法错误：/setcredit <UserId:Optional> <Credit>")
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
			SmartSendDelete(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.Credit))
		} else {
			SmartSendDelete(m, "❌ 您没有喵组权限，亦或是您未再对应群组使用这个命令")
		}
		LazyDelete(m)
	})

	Bot.Handle("/add_credit", func(m *tb.Message) {
		if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
			addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
			target := &CreditInfo{}
			credit := int64(0)

			if len(addons) == 0 {
				SmartSendDelete(m, "❌ 使用方法错误：/addcredit <UserId:Optional> <Credit>")
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
			SmartSendDelete(m, fmt.Sprintf("\u200d 设置成功，TA 的积分为: %d", target.Credit))
		} else {
			SmartSendDelete(m, "❌ 您没有喵组权限，亦或是您未再对应群组使用这个命令")
		}
		LazyDelete(m)
	})

	Bot.Handle("/set_channel", func(m *tb.Message) {
		if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
			gc := GetGroupConfig(m.Chat.ID)
			if gc != nil {
				if m.Payload == "" {
					gc.MustFollow = ""
					SetGroupConfig(m.Chat.ID, gc)
					SmartSendDelete(m, "\u200d 已经取消加群频道验证啦 ～")
				} else {
					if UserIsInGroup(m.Payload, Bot.Me.ID) != UIGIn {
						SmartSendDelete(m, "❌ 您还没有在辣个频道给我权限呢 TAT")
					} else {
						gc.MustFollow = m.Payload
						SetGroupConfig(m.Chat.ID, gc)
						SmartSendDelete(m, "\u200d 已经设置好加群频道验证啦 ～")
					}
				}
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/creditrank", func(m *tb.Message) {
		if IsGroupAdmin(m.Chat, m.Sender) {
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
			SmartSend(m, "👀 当前的积分墙为: \n\n"+rankStr, &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
			})
		} else {
			SmartSendDelete(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
		LazyDelete(m)
	})

	Bot.Handle("/lottery", func(m *tb.Message) {
		if IsGroupAdmin(m.Chat, m.Sender) {
			payloads := strings.Fields(m.Payload)

			rank, _ := strconv.Atoi(payloads[0])
			n, _ := strconv.Atoi(payloads[1])

			if rank <= 0 {
				rank = 10
			} else if rank > 100 {
				rank = 100
			}
			if n > rank {
				n = rank
			}

			ranks := GetCreditRank(m.Chat.ID, rank)
			sort.Slice(ranks, func(i, j int) bool {
				return rand.Intn(10) >= 5
			})
			rankStr := ""
			for i, c := range ranks[:n] {
				rankStr += fmt.Sprintf("`%2d.` `%s` ([%d](%s))\n", i+1, strings.ReplaceAll(c.Name, "`", "'"), c.ID, fmt.Sprintf("tg://user?id=%d", c.ID))
			}
			SmartSend(m, fmt.Sprintf("🎉 恭喜以下用户中奖：\n\n"+rankStr), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
			})
		} else {
			SmartSendDelete(m, "❌ 您没有权限，亦或是您未再对应群组使用这个命令")
		}
		LazyDelete(m)
	})

	// ---------------- Normal User ----------------

	Bot.Handle("/ban_user", func(m *tb.Message) {
		if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
			if err := Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
				SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 %s 获得禁言大礼包，可喜可贺可喜可贺！", GetUserName(m.ReplyTo.Sender)))
			} else {
				DErrorE(err, "Perm Update | Fail to ban user")
				SmartSendDelete(m, "❌ 您没有办法禁言 TA 呢")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/unban_user", func(m *tb.Message) {
		if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
			if err := Unban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
				SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 %s 重新获得了自由 ～", GetUserName(m.ReplyTo.Sender)))
			} else {
				DErrorE(err, "Perm Update | Fail to unban user")
				SmartSendDelete(m, "❌ 您没有办法解禁 TA 呢")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/kick_user", func(m *tb.Message) {
		if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
			if err := KickOnce(m.Chat.ID, m.ReplyTo.Sender.ID); err == nil {
				SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 %s 被踢出去啦！", GetUserName(m.ReplyTo.Sender)))
			} else {
				DErrorE(err, "Perm Update | Fail to kick user once")
				SmartSendDelete(m, "❌ 您没有踢掉 TA 呢")
			}
		}
		LazyDelete(m)
	})

	Bot.Handle("/mycredit", func(m *tb.Message) {
		if m.Chat.ID > 0 {
			SmartSendDelete(m, "❌ 请在群组发送这条命令来查看积分哦 ～")
		} else {
			SmartSendDelete(m, fmt.Sprintf("👀 您当前的积分为: %d", GetCredit(m.Chat.ID, m.Sender.ID).Credit))
		}
		LazyDelete(m)
	})

	Bot.Handle(tb.OnUserLeft, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if m.UserLeft.ID > 0 && !m.UserLeft.IsBot {
				gc := GetGroupConfig(m.Chat.ID)
				if gc != nil {
					gc.UpdateAdmin(m.UserLeft.ID, UMDel)
				}
				UpdateCredit(BuildCreditInfo(m.Chat.ID, m.UserLeft, false), UMDel, 0)
			}
		}
		LazyDelete(m)
	})

	// Bot.Handle("清除我的积分", func(m *tb.Message) {
	// 	if IsGroup(m.Chat.ID) {
	// 		UpdateCredit(BuildCreditInfo(m.Chat.ID, m.Sender, false), UMDel, 0)
	// 		SmartSendDelete(m, "好的")
	// 	}
	// 	LazyDelete(m)
	// })

	// Bot.Handle("频道测试", func(m *tb.Message) {
	// 	if gc := GetGroupConfig(m.Chat.ID); gc != nil && m.ReplyTo != nil && gc.MustFollow != "" {
	// 		i := UserIsInGroup(gc.MustFollow, m.ReplyTo.Sender.ID)
	// 		SmartSendDelete(m, fmt.Sprintf("状态：%v", i))
	// 	}
	// 	LazyDelete(m)
	// })

	Bot.Handle(tb.OnUserJoined, func(m *tb.Message) {
		CheckChannelFollow(m, m.UserJoined, true)
	})

	Bot.Handle(tb.OnPinned, func(m *tb.Message) {
		LazyDelete(m)
	})

	Bot.Handle("口臭", CMDWarnUser)
	Bot.Handle("口 臭", CMDWarnUser)
	Bot.Handle("嘴臭", CMDWarnUser)
	Bot.Handle("嘴 臭", CMDWarnUser)

	Bot.Handle("恶意广告", CMDBanUser)
	Bot.Handle("恶意发言", CMDBanUser)

	Bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		m := c.Message
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil {
			cmds := strings.Split(strings.TrimSpace(c.Data), "/")
			cmd, gid, uid, secuid := "", int64(0), int64(0), int64(0)
			if len(cmds) > 0 {
				cmd = cmds[0]
			}
			if len(cmds) > 1 {
				gid, _ = strconv.ParseInt(cmds[1], 10, 64)
			}
			if len(cmds) > 2 {
				uid, _ = strconv.ParseInt(cmds[2], 10, 64)
			}
			if len(cmds) > 3 {
				secuid, _ = strconv.ParseInt(cmds[3], 10, 64)
			}
			vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)
			isGroupAdmin := IsGroupAdmin(m.Chat, c.Sender)
			if strings.Contains("vt unban kick check", cmd) && IsGroup(gid) && uid > 0 {
				if cmd == "unban" && isGroupAdmin {
					if Unban(gid, uid, 0) == nil {
						Rsp(c, "✔️ 已解除封禁，请您手动处理后续事宜 ~")
					} else {
						Rsp(c, "❌ 解封失败，可能 TA 已经退群啦 ~")
					}
					SmartEdit(m, m.Text+"\n\nTA 已被管理员解封 👊")
					addCredit(gid, &tb.User{ID: uid}, 50, true)
					if secuid > 0 {
						votemap.Unset(vtToken)
						addCredit(gid, &tb.User{ID: secuid}, -15, true)
					}
				} else if cmd == "kick" && isGroupAdmin {
					if Kick(gid, uid) == nil {
						Rsp(c, "✔️ 已将 TA 送出群留学去啦 ~")
					} else {
						Rsp(c, "❌ 踢出失败，可能 TA 已经退群啦 ~")
					}
					votemap.Unset(vtToken)
					SmartEdit(m, m.Text+"\n\nTA 已被管理员踢出群聊 🦶")
				} else if cmd == "check" {
					if uid == c.Sender.ID {
						usrStatus := UserIsInGroup(gc.MustFollow, uid)
						if usrStatus == UIGIn {
							if Unban(gid, uid, 0) == nil {
								Bot.Delete(m)
								Rsp(c, "✔️ 验证成功，欢迎您的加入 ~")
							} else {
								Rsp(c, "❌ 验证成功，但是解禁失败，请联系管理员处理 ~")
							}
						} else {
							Rsp(c, "❌ 验证失败，请确认自己已经加入对应群组 ~")
						}
					} else {
						Rsp(c, "😠 人家的验证不要乱点哦！！！")
					}
				} else if cmd == "vt" {
					userVtToken := fmt.Sprintf("vu-%d,%d,%d", gid, uid, c.Sender.ID)
					if _, ok := votemap.Get(vtToken); ok {
						if votemap.Add(userVtToken) == 1 {
							votes := votemap.Add(vtToken)
							if votes >= 6 {
								Unban(gid, uid, 0)
								votemap.Unset(vtToken)
								SmartEdit(m, m.Text+"\n\n于多名用户投票后决定，该用户不是恶意广告，用户已解封，积分已原路返回。")
								addCredit(gid, &tb.User{ID: uid}, 50, true)
								if secuid > 0 {
									addCredit(gid, &tb.User{ID: secuid}, -15, true)
								}
							} else {
								EditBtns(m, m.Text, "", GenVMBtns(votes, gid, uid, secuid))
							}
							Rsp(c, "✔️ 投票成功，感谢您的参与 ~")
						} else {
							Rsp(c, "❌ 您已经参与过投票了，请不要多次投票哦 ~")
						}
					} else {
						Rsp(c, "❌ 投票时间已过，请联系管理员处理 ~")
					}
				} else {
					Rsp(c, "❌ 请不要乱玩管理员指令！")
				}
			} else {
				Rsp(c, "❌ 指令解析出错，请联系管理员解决 ~")
			}
		} else {
			Rsp(c, "❌ 这个群组还没有被授权哦 ~")
		}
	})

	Bot.Handle(tb.OnSticker, func(m *tb.Message) {
		CheckChannelFollow(m, m.Sender, false)
	})

	Bot.Handle(tb.OnPhoto, func(m *tb.Message) {
		CheckChannelFollow(m, m.Sender, false)
	})

	Bot.Handle(tb.OnDocument, func(m *tb.Message) {
		CheckChannelFollow(m, m.Sender, false)
	})

	Bot.Handle(tb.OnText, func(m *tb.Message) {
		if IsGroup(m.Chat.ID) {
			if !CheckChannelFollow(m, m.Sender, false) {
				return
			}

			if m.IsForwarded() {
				return
			}

			text := strings.TrimSpace(m.Text)
			textLen := len([]rune(text))
			userId := m.Sender.ID

			if m.Sender.Username != "Channel_Bot" {
				if puncReg.MatchString(text) {
					addCredit(m.Chat.ID, m.Sender, -5, true)
					lastID = userId
				} else if textLen >= 2 {
					if lastID == userId && text == lastText {
						addCredit(m.Chat.ID, m.Sender, -3, true)
					} else if lastID != userId || (textLen >= 14 && text != lastText) {
						addCredit(m.Chat.ID, m.Sender, 1, false)
					}
					lastID = userId
					lastText = text
				}
			}

			if ValidReplyUser(m) {
				addCredit(m.Chat.ID, m.ReplyTo.Sender, 1, true)
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
					addCredit(m.Chat.ID, m.Sender, 1, false)
					lastID = userId
				}
			}

			if ValidReplyUser(m) {
				addCredit(m.Chat.ID, m.ReplyTo.Sender, 1, true)
			}
		}
	})

	go Bot.Start()
	DInfo("MiaoKeeper is up.")
}

func CheckChannelFollow(m *tb.Message, user *tb.User, showExceptDialog bool) bool {
	if gc := GetGroupConfig(m.Chat.ID); gc != nil && gc.MustFollow != "" {
		usrName := strings.ReplaceAll(GetUserName(user), "`", "'")
		if user.IsBot {
			if showExceptDialog {
				SmartSendDelete(m.Chat, fmt.Sprintf("👏 欢迎 %s 加入群组，已为机器人自动放行 ～", usrName))
			}
			return true
		}
		usrStatus := UserIsInGroup(gc.MustFollow, user.ID)
		if usrStatus == UIGIn {
			if showExceptDialog {
				SmartSendDelete(m.Chat, fmt.Sprintf("👏 欢迎 %s 加入群组，您已关注频道自动放行 ～", usrName))
			}
		} else if usrStatus == UIGOut {
			chatId, userId := m.Chat.ID, user.ID
			msg, err := SendBtnsMarkdown(m.Chat, fmt.Sprintf("[🎉](tg://user?id=%d) 欢迎 `%s` 加入群组，您还没有关注本群组关联的频道哦，您有 5 分钟时间验证自己 ～ 请点击下面按钮跳转到频道关注后再回来验证以解除发言限制 ～", userId, usrName), "", []string{
				fmt.Sprintf("👉👉 跳转频道 👈👈|https://t.me/%s", strings.TrimLeft(gc.MustFollow, "@")),
				fmt.Sprintf("👉👉 点我验证 👈👈|check/%d/%d", chatId, userId),
				fmt.Sprintf("🚩 解封[管理]|unban/%d/%d||🚮 清退[管理]|kick/%d/%d", chatId, userId, chatId, userId),
			})
			if msg == nil || err != nil {
				if showExceptDialog {
					SmartSendDelete(m.Chat, "❌ 无法发送验证消息，请管理员检查群组权限 ～")
				}
			} else {
				if Ban(chatId, userId, 0) != nil {
					LazyDelete(msg)
					if showExceptDialog {
						SmartSendDelete(m.Chat, "❌ 无法完成验证流程，请管理员检查机器人封禁权限 ～")
					}
				} else {
					time.AfterFunc(time.Minute*5, func() {
						Bot.Delete(msg)
						cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
						if err != nil || cm.Role == tb.Restricted || cm.Role == tb.Kicked || cm.Role == tb.Left {
							Kick(chatId, userId)
							SmartSend(m.Chat, fmt.Sprintf("👀 [TA](tg://user?id=%d) 没有在规定时间内完成验证，已经被我带走啦 ～", userId), &tb.SendOptions{
								ParseMode:             "Markdown",
								DisableWebPagePreview: true,
							})
						}
					})
					return false
				}
			}
		} else {
			if showExceptDialog {
				SmartSendDelete(m.Chat, "❌ 无法检测用户是否在群组内，请管理员检查机器人权限 ～")
			}
		}
	}
	return true
}

func Rsp(c *tb.Callback, msg string) {
	Bot.Respond(c, &tb.CallbackResponse{
		Text:      msg,
		ShowAlert: true,
	})
}

func GenVMBtns(votes int, chatId, userId, secondUserId int64) []string {
	return []string{
		fmt.Sprintf("😠 这不公平 (%d)|vt/%d/%d/%d", votes, chatId, userId, secondUserId),
		fmt.Sprintf("🚩 解封[管理]|unban/%d/%d/%d||🚮 清退[管理]|kick/%d/%d/%d", chatId, userId, secondUserId, chatId, userId, secondUserId),
	}
}

func addCredit(chatId int64, user *tb.User, credit int64, force bool) *CreditInfo {
	if chatId < 0 && user != nil && user.ID > 0 && credit != 0 {
		token := fmt.Sprintf("ac-%d-%d", chatId, user.ID)
		if creditomap.Add(token) < 20 || force { // can only get credit 20 times / hour
			return UpdateCredit(BuildCreditInfo(chatId, user, false), UMAdd, credit)
		}
	}
	return nil
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

func SmartEdit(to *tb.Message, what interface{}, options ...interface{}) (*tb.Message, error) {
	options = append([]interface{}{&tb.SendOptions{
		// ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	}}, options...)
	m, err := Bot.Edit(to, what, options...)
	if err != nil {
		DErrorE(err, "Telegram Edit Error")
	}
	return m, err
}

func SmartSendDelete(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	msg, err := SmartSend(to, what, options...)
	if err == nil && msg != nil {
		LazyDelete(msg)
	}
	return msg, err
}

func MakeBtns(prefix string, btns []string) [][]tb.InlineButton {
	btnsc := make([][]tb.InlineButton, 0)
	for _, row := range btns {
		btnscr := make([]tb.InlineButton, 0)
		for _, btn := range strings.Split(row, "||") {
			z := strings.SplitN(btn, "|", 2)
			if len(z) < 2 {
				continue
			}
			unique := ""
			link := ""
			if _, err := url.Parse(z[1]); err == nil && strings.HasPrefix(z[1], "https://") {
				link = z[1]
			} else {
				unique = prefix + z[1]
			}
			btnscr = append(btnscr, tb.InlineButton{
				Unique: unique,
				Text:   z[0],
				Data:   "",
				URL:    link,
			})
		}
		btnsc = append(btnsc, btnscr)
	}
	return btnsc
}

func SendBtns(to interface{}, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartSendInner(to, what, &tb.SendOptions{
		// ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	}, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          true,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func SendBtnsMarkdown(to interface{}, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartSendInner(to, what, &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	}, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          true,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func EditBtns(to *tb.Message, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartEdit(to, what, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          true,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func SmartSend(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	if len(options) == 0 {
		return SmartSendInner(to, what, &tb.SendOptions{
			// ParseMode:             "Markdown",
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
	}

	return s
}

func GetChatName(u *tb.Chat) string {
	s := ""
	if u.FirstName != "" || u.LastName != "" {
		s = strings.TrimSpace(u.FirstName + " " + u.LastName)
	} else if u.Username != "" {
		s = "@" + u.Username
	}

	return s
}

func UserIsInGroup(chatRepr string, userId int64) UIGStatus {
	cm, err := ChatMemberOf(chatRepr, Bot.Me.ID)
	if err != nil {
		return UIGErr
	} else if cm.Role != tb.Administrator && cm.Role != tb.Creator {
		return UIGErr
	}

	if userId == Bot.Me.ID {
		return UIGIn
	}

	cm, err = ChatMemberOf(chatRepr, userId)
	if err != nil || cm == nil {
		return UIGOut
	}
	if cm.Role == tb.Left || cm.Role == tb.Kicked {
		return UIGOut
	}
	return UIGIn
}

func ChatMemberOf(chatRepr string, userId int64) (*tb.ChatMember, error) {
	params := map[string]string{
		"chat_id": chatRepr,
		"user_id": strconv.FormatInt(userId, 10),
	}

	data, err := Bot.Raw("getChatMember", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result *tb.ChatMember
	}
	if err := jsoniter.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func Kick(chatId, userId int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		return Bot.Ban(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func KickOnce(chatId, userId int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		err = Bot.Ban(&tb.Chat{ID: chatId}, cm)
		if err == nil {
			return Bot.Unban(&tb.Chat{ID: chatId}, &tb.User{ID: userId}, true)
		}
	}
	return err
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

func BanChannel(chatId, channelId int64) error {
	params := map[string]interface{}{
		"chat_id":        strconv.FormatInt(chatId, 10),
		"sender_chat_id": strconv.FormatInt(channelId, 10),
	}

	_, err := Bot.Raw("banChatSenderChat", params)
	return err
}

func init() {
	puncReg = regexp.MustCompile(`^[!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`" + `][a-zA-Z0-9]+`)
	zcomap = NewOMap(60*60*1000, true)
	creditomap = NewOMap(60*60*1000, false)
	votemap = NewOMap(30*60*1000, false)
}
