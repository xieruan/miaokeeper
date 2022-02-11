package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdOnText(m *tb.Message) {
	if gc := GetGroupConfig(m.Chat.ID); gc != nil {
		if !CheckChannelForward(m) {
			return
		}

		if !CheckChannelFollow(m, m.Sender, false) {
			return
		}

		if rule := gc.TestCustomReplyRule(m); rule != nil {
			if rule.CreditBehavior != 0 {
				addCreditToMsgSender(m.Chat.ID, m, int64(rule.CreditBehavior), true)
			}

			if rule.ReplyMessage != "" {
				var target interface{} = m
				if rule.ReplyTo == "group" {
					target = m.Chat
				} else if rule.ReplyTo == "private" {
					target = m.Sender
				}

				_, err := SmartSendWithBtns(target, BuilRuleMessage(rule.ReplyMessage, m), BuildRuleMessages(rule.ReplyButtons, m), &tb.SendOptions{
					ParseMode:             "Markdown",
					DisableWebPagePreview: true,
					AllowWithoutReply:     true,
				})

				if err != nil {
					SmartSendDelete(m, Locale("system.notsend", GetSenderLocale(m))+"\n\n"+err.Error())
				}
			}

			if APIToken != "" && rule.CallbackURL != "" {
				if u, err := url.Parse(rule.CallbackURL); err == nil && u != nil {
					go POSTJsonWithSign(rule.CallbackURL, []byte(rule.ToJson(false)), time.Second*3)
				}
			}

			return
		}

		if CheckSpoiler(m) {
			RevealSpoiler(m)
			addCreditToMsgSender(m.Chat.ID, m, -2, true)
			return
		}

		if m.IsForwarded() {
			return
		}

		if gc.IsBanKeyword(m) {
			CmdBanUser(m)
			return
		} else if gc.IsWarnKeyword(m) {
			CmdWarnUser(m)
			return
		}

		text := strings.TrimSpace(m.Text)
		textLen := len([]rune(text))
		userId := m.Sender.ID

		if puncReg.MatchString(text) {
			// commands
			addCreditToMsgSender(m.Chat.ID, m, -5, true)
		} else if lastID == userId && text == lastText {
			// duplicated messages
			addCreditToMsgSender(m.Chat.ID, m, -2, true)
		} else if textLen >= 2 && (lastID != userId || (textLen >= 14 && text != lastText)) {
			// valid messages
			addCreditToMsgSender(m.Chat.ID, m, 1, false)
			if ValidReplyUser(m) {
				addCreditToMsgSender(m.Chat.ID, m.ReplyTo, 1, true)
			}
		}

		lastID = userId
		lastText = text
	}
}

func CmdOnSticker(m *tb.Message) {
	if IsGroup(m.Chat.ID) {
		if !CheckChannelForward(m) {
			return
		}
		if !CheckChannelFollow(m, m.Sender, false) {
			return
		}

		if m.IsForwarded() {
			return
		}
		userId := m.Sender.ID
		if lastID != userId {
			addCreditToMsgSender(m.Chat.ID, m, 1, false)
			lastID = userId
		}

		if ValidReplyUser(m) {
			addCreditToMsgSender(m.Chat.ID, m.ReplyTo, 1, true)
		}
	}
}

func CmdOnDocument(m *tb.Message) {
	if m.Caption == "/su_import_credit" && m.Document != nil {
		CmdSuImportCredit(m)
	} else if m.Caption == "/import_policy" && m.Document != nil {
		CmdImportPolicy(m)
	} else {
		CheckChannelForward(m)
		CheckChannelFollow(m, m.Sender, false)
	}
}

func CmdOnUserLeft(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && m.UserLeft.ID > 0 {
		gc.UpdateAdmin(m.UserLeft.ID, UMDel)
		UpdateCredit(BuildCreditInfo(m.Chat.ID, m.UserLeft, false), UMDel, 0)
	}
	LazyDelete(m)
}

func CmdOnChatMember(cmu *tb.ChatMemberUpdated) {
	gc := GetGroupConfig(cmu.Chat.ID)
	if gc != nil && cmu.NewChatMember != nil && cmu.NewChatMember.User != nil && cmu.NewChatMember.User.ID > 0 {
		user := cmu.NewChatMember.User
		if cmu.NewChatMember.Role == tb.Kicked ||
			cmu.NewChatMember.Role == tb.Left {
			gc.UpdateAdmin(user.ID, UMDel)
			UpdateCredit(BuildCreditInfo(cmu.Chat.ID, user, false), UMDel, 0)
		}
	}
}

func CmdOnUserJoined(m *tb.Message) {
	if gc := GetGroupConfig(m.Chat.ID); gc != nil {
		if gc.IsBlackListName(m.Sender) {
			KickOnce(m.Chat.ID, m.Sender.ID)
			SmartSend(m.Chat, fmt.Sprintf(Locale("channel.pattern.kicked", GetSenderLocale(m)), m.Sender.ID), &tb.SendOptions{
				ParseMode:             "Markdown",
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			})
			return
		}
	}
	CheckChannelFollow(m, m.UserJoined, true)
}

func CmdOnPinned(m *tb.Message) {
	LazyDelete(m)
}

func CmdOnMisc(m *tb.Message) {
	CheckChannelForward(m)
	CheckChannelFollow(m, m.Sender, false)
}
