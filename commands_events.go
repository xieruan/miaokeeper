package main

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdOnText(m *tb.Message) {
	if IsGroup(m.Chat.ID) {
		if !CheckChannelForward(m) {
			return
		}

		if !CheckChannelFollow(m, m.Sender, false) {
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

		text := strings.TrimSpace(m.Text)
		textLen := len([]rune(text))
		userId := m.Sender.ID

		if puncReg.MatchString(text) {
			addCreditToMsgSender(m.Chat.ID, m, -5, true)
			lastID = userId
		} else if textLen >= 2 {
			if lastID == userId && text == lastText {
				addCreditToMsgSender(m.Chat.ID, m, -2, true)
			} else if lastID != userId || (textLen >= 14 && text != lastText) {
				addCreditToMsgSender(m.Chat.ID, m, 1, false)
			}
			lastID = userId
			lastText = text
		}

		if ValidReplyUser(m) {
			addCreditToMsgSender(m.Chat.ID, m.ReplyTo, 1, true)
		}
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
	CheckChannelFollow(m, m.UserJoined, true)
}

func CmdOnPinned(m *tb.Message) {
	LazyDelete(m)
}

func CmdOnMisc(m *tb.Message) {
	CheckChannelForward(m)
	CheckChannelFollow(m, m.Sender, false)
}
