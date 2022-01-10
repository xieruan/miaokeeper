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
		SmartSendDelete(m, fmt.Sprintf(Locale("credit.importSuccess", m.Sender.LanguageCode), len(records)))
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
				SmartSendDelete(m.ReplyTo, Locale("grant.assign.success", m.Sender.LanguageCode))
			} else {
				SmartSendDelete(m.ReplyTo, Locale("grant.assign.failure", m.Sender.LanguageCode))
			}
		} else {
			SmartSendDelete(m, Locale("cmd.mustReply", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdSuDelAdmin(m *tb.Message) {
	if IsAdmin(m.Sender.ID) {
		if ValidMessageUser(m.ReplyTo) {
			if UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
				SmartSendDelete(m.ReplyTo, Locale("grant.dismiss.success", m.Sender.LanguageCode))
			} else {
				SmartSendDelete(m.ReplyTo, Locale("grant.dismiss.failure", m.Sender.LanguageCode))
			}
		} else {
			SmartSendDelete(m, Locale("cmd.mustReply", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

// ---------------- Group Admin ----------------

func CmdAddAdmin(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) && m.ReplyTo != nil {
		if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMAdd) {
			SmartSendDelete(m.ReplyTo, Locale("grant.assign.success", m.Sender.LanguageCode))
		} else {
			SmartSendDelete(m.ReplyTo, Locale("grant.assign.failure", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdDelAdmin(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && (gc.IsAdmin(m.Sender.ID) || IsAdmin(m.Sender.ID)) && m.ReplyTo != nil {
		if gc.UpdateAdmin(m.ReplyTo.Sender.ID, UMDel) {
			SmartSendDelete(m.ReplyTo, Locale("grant.dismiss.success", m.Sender.LanguageCode))
		} else {
			SmartSendDelete(m.ReplyTo, Locale("grant.dismiss.failure", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
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
				SmartSendDelete(m, Locale("forward.ban.success", m.Sender.LanguageCode))
			} else {
				SmartSendDelete(m, Locale("forward.ban.failure", m.Sender.LanguageCode))
			}
		} else {
			SmartSendDelete(m, Locale("cmd.mustReplyChannelOrInput", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
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
				SmartSendDelete(m, Locale("forward.unban.success", m.Sender.LanguageCode))
			} else {
				SmartSendDelete(m, Locale("forward.unban.failure", m.Sender.LanguageCode))
			}
		} else {
			SmartSendDelete(m, Locale("cmd.mustReplyChannelOrInput", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noGroupPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdSetCredit(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
		target := &CreditInfo{}
		credit := int64(0)

		if len(addons) == 0 {
			SmartSendDelete(m, Locale("credit.set.invalid", m.Sender.LanguageCode))
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
		SmartSendDelete(m, fmt.Sprintf(Locale("credit.set.success", m.Sender.LanguageCode), target.Credit))
	} else {
		SmartSendDelete(m, Locale("cmd.noMiaoPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdAddCredit(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		addons := ParseStrToInt64Arr(strings.Join(strings.Fields(strings.TrimSpace(m.Payload)), ","))
		target := &CreditInfo{}
		credit := int64(0)

		if len(addons) == 0 {
			SmartSendDelete(m, Locale("credit.add.invalid", m.Sender.LanguageCode))
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
		SmartSendDelete(m, fmt.Sprintf(Locale("credit.set.success", m.Sender.LanguageCode), target.Credit))
	} else {
		SmartSendDelete(m, Locale("cmd.noMiaoPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdCheckCredit(m *tb.Message) {
	if IsGroupAdminMiaoKo(m.Chat, m.Sender) {
		if m.Chat.ID > 0 || !m.IsReply() {
			SmartSendDelete(m, Locale("cmd.mustReply", m.Sender.LanguageCode))
		} else {
			SmartSendDelete(m, fmt.Sprintf(Locale("credit.check.success", m.Sender.LanguageCode), GetQuotableUserName(m.ReplyTo.Sender), GetCredit(m.Chat.ID, m.ReplyTo.Sender.ID).Credit), &tb.SendOptions{
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
				SmartSendDelete(m, Locale("spoiler.invalid", m.Sender.LanguageCode))
				LazyDelete(m)
				return
			}

			gc.AntiSpoiler = status
			SetGroupConfig(m.Chat.ID, gc)
			SmartSendDelete(m, fmt.Sprintf(Locale("spoiler.success", m.Sender.LanguageCode), gc.AntiSpoiler), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noMiaoPerm", m.Sender.LanguageCode))
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
				SmartSendDelete(m, Locale("channel.set.cancel", m.Sender.LanguageCode))
			} else {
				if UserIsInGroup(groupName, Bot.Me.ID) != UIGIn {
					SmartSendDelete(m, Locale("channel.cannotCheckChannel", m.Sender.LanguageCode))
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
					SmartSendDelete(m, fmt.Sprintf(Locale("channel.set.success", m.Sender.LanguageCode), gc.MustFollowOnJoin, gc.MustFollowOnMsg), &tb.SendOptions{
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
						AllowWithoutReply:     true,
					})
				}
			}
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noMiaoPerm", m.Sender.LanguageCode))
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
		redpacketrankmap.Set(redpacketKey+":sender", Locale("rp.admin", m.Sender.LanguageCode)+GetQuotableUserName(m.Sender))
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
		SmartSend(m, Locale("credit.rank.info", m.Sender.LanguageCode)+rankStr, &tb.SendOptions{
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
			SmartSendDelete(m, Locale("system.unexpected", m.Sender.LanguageCode))
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
			SmartSendDelete(m, Locale("rp.set.invalid", m.Sender.LanguageCode), &tb.SendOptions{
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
			SmartSendDelete(m, Locale("rp.set.noEnoughCredit", m.Sender.LanguageCode))
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
		SmartSend(m, fmt.Sprintf(Locale("credit.lottery.info", m.Sender.LanguageCode)+rankStr), &tb.SendOptions{
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
			SmartSendDelete(m, fmt.Sprintf(Locale("gp.ban.success", m.Sender.LanguageCode), GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to ban user")
			SmartSendDelete(m, Locale("gp.ban.failure", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdUnbanUserCommand(m *tb.Message) {
	if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
		if err := Unban(m.Chat.ID, m.ReplyTo.Sender.ID, 0); err == nil {
			SmartSendDelete(m, fmt.Sprintf(Locale("gp.unban.success", m.Sender.LanguageCode), GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to unban user")
			SmartSendDelete(m, Locale("gp.unban.failure", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdKickUserCommand(m *tb.Message) {
	if IsGroupAdmin(m.Chat, m.Sender) && ValidReplyUser(m) {
		if err := KickOnce(m.Chat.ID, m.ReplyTo.Sender.ID); err == nil {
			SmartSendDelete(m, fmt.Sprintf(Locale("gp.kick.success", m.Sender.LanguageCode), GetQuotableUserName(m.ReplyTo.Sender)), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
		} else {
			DErrorE(err, "Perm Update | Fail to kick user once")
			SmartSendDelete(m, Locale("gp.kick.failure", m.Sender.LanguageCode))
		}
	} else {
		SmartSendDelete(m, Locale("cmd.noPerm", m.Sender.LanguageCode))
	}
	LazyDelete(m)
}

func CmdMyCredit(m *tb.Message) {
	if m.Chat.ID > 0 {
		SmartSendDelete(m, Locale("cmd.mustInGroup", m.Sender.LanguageCode))
	} else if IsGroup(m.Chat.ID) {
		SmartSendDelete(m, fmt.Sprintf(Locale("credit.check.my", m.Sender.LanguageCode), GetQuotableUserName(m.Sender), GetCredit(m.Chat.ID, m.Sender.ID).Credit), &tb.SendOptions{
			ParseMode:             "Markdown",
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		})
	}
	LazyDelete(m)
}

func CmdVersion(m *tb.Message) {
	SmartSendDelete(m, fmt.Sprintf(Locale("cmd.misc.version", m.Sender.LanguageCode), VERSION))
	LazyDelete(m)
}

func CmdPing(m *tb.Message) {
	t := time.Now().UnixMilli()
	Bot.GetCommands()
	t1 := time.Now().UnixMilli() - t
	msg, _ := SmartSendDelete(m.Chat, fmt.Sprintf(Locale("cmd.misc.ping.1", m.Sender.LanguageCode), t1), &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	})
	t2 := time.Now().UnixMilli() - t - t1
	SmartEdit(msg, fmt.Sprintf(Locale("cmd.misc.ping.2", m.Sender.LanguageCode), t1, t2), &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	})
	LazyDelete(m)
}
