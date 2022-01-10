package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdSuExportCredit(m *tb.Message) {
	defer LazyDelete(m)
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && IsAdmin(m.Sender.ID) {
		err := Bot.Notify(m.Sender, tb.UploadingDocument)
		if err != nil {
			SmartSendDelete(m, Locale("cmd.privateChatFirst", m.Sender.LanguageCode))
			return
		}
		records := DumpCredits(m.Chat.ID)
		ioBuffer := bytes.Buffer{}
		w := csv.NewWriter(&ioBuffer)
		w.WriteAll(records)
		Bot.Send(m.Sender, &tb.Document{
			File:     tb.FromReader(&ioBuffer),
			MIME:     "text/csv",
			FileName: "CreditDump" + time.Now().Format(time.RFC3339) + ".csv",
		})
		SmartSendDelete(m, Locale("credit.exportSuccess", m.Sender.LanguageCode))
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
}

func CmdSuImportCredit(m *tb.Message) {
	Bot.Delete(m)
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && IsAdmin(m.Sender.ID) {
		Bot.Notify(m.Chat, tb.UploadingDocument)
		ioHandler, err := Bot.GetFile(&m.Document.File)
		if err != nil {
			SmartSendDelete(m, Locale("credit.importError", m.Sender.LanguageCode))
			DErrorEf(err, "Import Credit Error | not downloaded url=%s", Bot.URL+"/file/bot"+Bot.Token+"/"+m.Document.FilePath)
			return
		}
		csvHandler := csv.NewReader(ioHandler)
		records, err := csvHandler.ReadAll()
		if err != nil {
			SmartSendDelete(m, Locale("credit.importParseError", m.Sender.LanguageCode))
			DErrorE(err, "Import Credit Error | not parsed")
			return
		}
		FlushCredits(m.Chat.ID, records)
		SmartSendDelete(m, fmt.Sprintf("\u200d 导入 %d 条成功，您可以输入 /creditrank 查看导入后积分详情", len(records)))
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
}

func CmdSuAddGroup(m *tb.Message) {
	if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
		if UpdateGroup(m.Chat.ID, UMAdd) {
			SmartSendDelete(m, Locale("su.group.addSuccess", m.Sender.LanguageCode))
		} else {
			SmartSendDelete(m, Locale("su.group.addDuplicate", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdSuDelGroup(m *tb.Message) {
	if IsAdmin(m.Sender.ID) && m.Chat.ID < 0 {
		if UpdateGroup(m.Chat.ID, UMDel) {
			SmartSendDelete(m, Locale("su.group.delSuccess", m.Sender.LanguageCode))
		} else {
			SmartSendDelete(m, Locale("su.group.delDuplicate", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdSuAddAdmin(m *tb.Message) {
	if IsAdmin(m.Sender.ID) {
		if ValidMessageUser(m.ReplyTo) {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
				SmartSendDelete(m.ReplyTo, "✔️ TA 已经成为管理员啦 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 已经是管理员啦 ～")
			}
		} else {
			SmartSendDelete(m, "❌ 请在群组内回复一个有效用户使用这个命令哦 ～")
		}
	} else {
		SmartSendDelete(m, "❌ 您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdSuDelAdmin(m *tb.Message) {
	if IsAdmin(m.Sender.ID) {
		if ValidMessageUser(m.ReplyTo) {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSendDelete(m.ReplyTo, "✔️ 已将 TA 的管理员移除 ～")
			} else {
				SmartSendDelete(m.ReplyTo, "❌ TA 本来就不是管理员呢 ～")
			}
		} else {
			SmartSendDelete(m, "❌ 请在群组内回复一个有效用户使用这个命令哦 ～")
		}
	} else {
		SmartSendDelete(m, "❌ 您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

// ---------------- Group Admin ----------------

func CmdAddAdmin(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
		if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
			SmartSendDelete(m.ReplyTo, "✔️ TA 已经成为群管理员啦 ～")
		} else {
			SmartSendDelete(m.ReplyTo, "❌ TA 已经是群管理员啦 ～")
		}
	} else {
		SmartSendDelete(m, "❌ 当前群组没有开启统计，或是您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdDelAdmin(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
		if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
			SmartSendDelete(m.ReplyTo, "✔️ 已将 TA 的群管理员移除 ～")
		} else {
			SmartSendDelete(m.ReplyTo, "❌ TA 本来就不是群管理员呢 ～")
		}
	} else {
		SmartSendDelete(m, "❌ 当前群组没有开启统计，或是您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdBanForward(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
		isReply := false
		id, _ := strconv.ParseInt(m.Payload, 10, 64)
		if id == 0 && m.IsReply() && m.ReplyTo.IsForwarded() && m.ReplyTo.OriginalChat != nil {
			id = m.ReplyTo.OriginalChat.ID
			isReply = true
		}
		if id != 0 {
			if gc.UpdateBannedForward(id, UMAdd) {
				if isReply {
					Bot.Delete(m.ReplyTo)
				}
				SmartSendDelete(m, "✔️ TA 已经被我封掉啦 ～")
			} else {
				SmartSendDelete(m, "❌ TA 已经被封禁过啦 ～")
			}
		} else {
			SmartSendDelete(m, "❌ 错误的使用方式，请回复一则转发的频道消息或者手动加上频道 id ～")
		}
	} else {
		SmartSendDelete(m, "❌ 当前群组没有开启统计，或是您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdUnbanForward(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) {
		id, _ := strconv.ParseInt(m.Payload, 10, 64)
		if id == 0 && m.IsReply() && m.ReplyTo.IsForwarded() && m.ReplyTo.OriginalChat != nil {
			id = m.ReplyTo.OriginalChat.ID
		}
		if id != 0 {
			if gc.UpdateBannedForward(id, UMDel) {
				SmartSendDelete(m, "✔️ TA 已经被我解封啦 ～")
			} else {
				SmartSendDelete(m, "❌ TA 还没有被封禁哦 ～")
			}
		} else {
			SmartSendDelete(m, "❌ 错误的使用方式，请回复一则转发的频道消息或者手动加上频道 id ～")
		}
	} else {
		SmartSendDelete(m, "❌ 当前群组没有开启统计，或是您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdSetCredit(m *tb.Message) {
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
}

func CmdAddCredit(m *tb.Message) {
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
}

func CmdCheckCredit(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		if m.Chat.ID > 0 {
			SmartSendDelete(m, "❌ 请在群组回复一个用户这条命令来查询 TA 的积分哦 ～")
		} else if !m.IsReply() {
			SmartSendDelete(m, "❌ 请回复一个用户这条命令来查询 TA 的积分哦 ～")
		} else {
			SmartSendDelete(m, fmt.Sprintf("👀 `%s`, TA 当前的积分为: %d", GetQuotableUserName(m.ReplyTo.Sender), GetCredit(m.Chat.ID, m.ReplyTo.Sender.ID).Credit), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdSetAntiSpoiler(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil {
			status := false

			if m.Payload == "on" {
				status = true
			} else if m.Payload == "off" {
				status = false
			} else {
				SmartSendDelete(m, "❌ 使用方法错误：/set_antispoiler <on|off>")
				LazyDelete(m)
				return
			}

			gc.AntiSpoiler = status
			SetGroupConfig(m.Chat.ID, gc)
			SmartSendDelete(m, fmt.Sprintf("\u200d 已经设置好反·反剧透消息啦 `(Status=%v)` ～", gc.AntiSpoiler), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		}
	} else {
		SmartSendDelete(m, "❌ 您没有喵组权限，亦或是您未再对应群组使用这个命令")
	}
	LazyDelete(m)
}

func CmdSetChannel(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		gc := GetGroupConfig(m.Chat.ID)
		if gc != nil {
			payloads := strings.Fields(strings.TrimSpace(m.Payload))
			groupName := ""
			mode := ""
			if len(payloads) > 0 {
				groupName = payloads[0]
			}
			if len(payloads) > 1 {
				mode = payloads[1]
			}
			if groupName == "" {
				gc.MustFollow = ""
				gc.MustFollowOnJoin = false
				gc.MustFollowOnMsg = false
				SetGroupConfig(m.Chat.ID, gc)
				SmartSendDelete(m, "\u200d 已经取消加群频道验证啦 ～")
			} else {
				if UserIsInGroup(groupName, Bot.Me.ID) != UIGIn {
					SmartSendDelete(m, "❌ 您还没有在辣个频道给我权限呢 TAT")
				} else {
					gc.MustFollow = groupName
					gc.MustFollowOnJoin = false
					gc.MustFollowOnMsg = false
					if mode == "join" {
						gc.MustFollowOnJoin = true
					} else if mode == "msg" {
						gc.MustFollowOnMsg = true
					} else {
						gc.MustFollowOnJoin = true
						gc.MustFollowOnMsg = true
					}
					SetGroupConfig(m.Chat.ID, gc)
					SmartSendDelete(m, fmt.Sprintf("\u200d 已经设置好加群频道验证啦 `(Join=%v, Msg=%v)` ～", gc.MustFollowOnJoin, gc.MustFollowOnMsg), &tb.SendOptions{
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
						AllowWithoutReply:     true,
					})
				}
			}
		}
	} else {
		SmartSendDelete(m, "❌ 您没有喵组权限，亦或是您未再对应群组使用这个命令")
	}
	LazyDelete(m)
}

func CmdSendRedpacket(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		payloads := strings.Fields(m.Payload)

		mc := 0
		if len(payloads) > 0 {
			mc, _ = strconv.Atoi(payloads[0])
		}
		n := 0
		if len(payloads) > 1 {
			n, _ = strconv.Atoi(payloads[1])
		}

		if mc <= 0 {
			mc = 1
		} else if mc > 1000000 {
			mc = 1000000
		}
		if n < 1 {
			n = 1
		} else if n > 1000 {
			n = 1000
		}

		chatId := m.Chat.ID
		redpacketId := time.Now().Unix() + int64(rand.Intn(10000))
		redpacketKey := fmt.Sprintf("%d-%d", chatId, redpacketId)
		redpacketrankmap.Set(redpacketKey+":sender", "管理员-"+GetQuotableUserName(m.Sender))
		redpacketmap.Set(redpacketKey, mc)
		redpacketnmap.Set(redpacketKey, n)
		SendRedPacket(m.Chat, chatId, redpacketId)
		LazyDelete(m)
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdCreditRank(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		rank, _ := strconv.Atoi(m.Payload)
		if rank <= 0 {
			rank = 10
		} else if rank > 50 {
			rank = 50
		}
		ranks := GetCreditRank(m.Chat.ID, rank)
		rankStr := ""
		for i, c := range ranks {
			rankStr += fmt.Sprintf("`%2d`. `%s`: `%d`\n", i+1, strings.ReplaceAll(c.Name, "`", "'"), c.Credit)
		}
		SmartSend(m, "#开榜 当前的积分墙为: \n\n"+rankStr, &tb.SendOptions{
			ParseMode:             "Markdown",
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		})
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdCreateLottery(m *tb.Message) {
	// :limit=(0-inf)
	// :consume=n|y
	// :num=1|100
	// :draw=manual|>num
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		payload, ah := ArgParse(m.Payload)
		limit, _ := ah.Int("limit")
		consume, _ := ah.Bool("consume")
		num, _ := ah.Int("num")
		if num <= 0 || num >= 100 {
			num = 1
		}
		duration, _ := ah.Int("duration")
		if duration <= 0 || duration >= 72 {
			duration = 0
		}
		participant, _ := ah.Int("participant")
		if participant < num {
			participant = 0
		}

		li := CreateLottery(m.Chat.ID, payload, limit, consume, num, duration, participant)

		if li != nil {
			li.UpdateTelegramMsg()
		} else {
			SmartSendDelete(m, "❌ 无法创建抽奖任务，请检查服务器错误日志")
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdRedpacket(m *tb.Message) {
	if IsGroup(m.Chat.ID) {
		payloads := strings.Fields(m.Payload)

		mc := 0
		if len(payloads) > 0 {
			mc, _ = strconv.Atoi(payloads[0])
		}
		n := 0
		if len(payloads) > 1 {
			n, _ = strconv.Atoi(payloads[1])
		}

		if mc <= 0 || n <= 0 || mc > 1000 || n > 20 || mc < n {
			SmartSendDelete(m, "❌ 使用方法不正确呢，请输入 /redpacket `<总分数>` `<红包个数>` 来发红包哦～\n\n备注：红包总分需在 1 ~ 1000 之间，红包个数需在 1 ~ 20 之间，且红包大小不能低于参与人数哦～", &tb.SendOptions{
				ParseMode: "Markdown",
			})
			LazyDelete(m)
			return
		}

		userredpacketlock.Lock()
		defer userredpacketlock.Unlock()
		ci := GetCredit(m.Chat.ID, m.Sender.ID)

		if ci != nil && ci.Credit >= int64(mc) {
			chatId := m.Chat.ID
			addCredit(chatId, m.Sender, -Abs(int64(mc)), true)
			redpacketId := time.Now().Unix() + int64(rand.Intn(10000))
			redpacketKey := fmt.Sprintf("%d-%d", chatId, redpacketId)
			redpacketrankmap.Set(redpacketKey+":sender", GetQuotableUserName(m.Sender))
			redpacketmap.Set(redpacketKey, mc)
			redpacketnmap.Set(redpacketKey, n)
			SendRedPacket(m.Chat, chatId, redpacketId)
		} else {
			SmartSendDelete(m, "❌ 您的积分不够发这个红包哦，请在努力赚积分吧～")
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdLottery(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		payloads := strings.Fields(m.Payload)

		rank := 0
		if len(payloads) > 0 {
			rank, _ = strconv.Atoi(payloads[0])
		}
		n := 0
		if len(payloads) > 1 {
			n, _ = strconv.Atoi(payloads[1])
		}

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
			AllowWithoutReply:     true,
		})
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

// ---------------- Normal User ----------------

func CmdBanUserCommand(m *tb.Message) {
	if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
		if err := Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
			SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 `%s` 获得禁言大礼包，可喜可贺可喜可贺！", GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to ban user")
			SmartSendDelete(m, "❌ 您没有办法禁言 TA 呢")
		}
	} else {
		SmartSendDelete(m, "❌ 您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdUnbanUserCommand(m *tb.Message) {
	if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
		if err := Unban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
			SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 `%s` 重新获得了自由 ～", GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to unban user")
			SmartSendDelete(m, "❌ 您没有办法解禁 TA 呢")
		}
	} else {
		SmartSendDelete(m, "❌ 您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdKickUserCommand(m *tb.Message) {
	if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
		if err := KickOnce(m.Chat.ID, m.ReplyTo.Sender.ID); err == nil {
			SmartSendDelete(m, fmt.Sprintf("🎉 恭喜 `%s` 被踢出去啦！", GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to kick user once")
			SmartSendDelete(m, "❌ 您没有踢掉 TA 呢")
		}
	} else {
		SmartSendDelete(m, "❌ 您没有使用这个命令的权限呢")
	}
	LazyDelete(m)
}

func CmdMyCredit(m *tb.Message) {
	if m.Chat.ID > 0 {
		SmartSendDelete(m, "❌ 请在群组发送这条命令来查看积分哦 ～")
	} else if IsGroup(m.Chat.ID) {
		SmartSendDelete(m, fmt.Sprintf("👀 `%s`, 您当前的积分为: %d", GetQuotableUserName(m.Sender), GetCredit(m.Chat.ID, m.Sender.ID).Credit), &tb.SendOptions{
			ParseMode:             "Markdown",
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		})
	}
	LazyDelete(m)
}

func CmdVersion(m *tb.Message) {
	SmartSendDelete(m, fmt.Sprintf("👀 当前版本为: %s", VERSION))
	LazyDelete(m)
}

func CmdPing(m *tb.Message) {
	t := time.Now().UnixMilli()
	Bot.GetCommands()
	t1 := time.Now().UnixMilli() - t
	msg, _ := SmartSendDelete(m.Chat, fmt.Sprintf("🔗 与 Telegram 伺服器的延迟约为:\n\n机器人 DC: `%dms`", t1), &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	})
	t2 := time.Now().UnixMilli() - t - t1
	SmartEdit(msg, fmt.Sprintf("🔗 与 Telegram 伺服器的延迟约为:\n\n机器人 DC: `%dms`\n群组 DC: `%dms`", t1, t2), &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	})
	LazyDelete(m)
}
