package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CMDWarnUser(m *tb.Message) {
	if IsGroup(m.Chat.ID) && m.ReplyTo != nil {
		if m.Sender.ID > 0 && m.Sender.Username != "Channel_Bot" {
			if m.ReplyTo.Sender.ID == m.Sender.ID {
				SmartSend(m, "确实")
			} else if m.ReplyTo.Sender.ID < 0 || m.ReplyTo.Sender.IsBot {
				SmartSend(m, "我拿它没办法呢 ...")
			} else {
				token := fmt.Sprintf("%d,%d,%d", m.Chat.ID, m.Sender.ID, m.ReplyTo.Sender.ID)
				limSenderToken := fmt.Sprintf("lim%d,%d,%d", m.Chat.ID, m.Sender.ID)
				limReciverToken := fmt.Sprintf("lim%d,%d,%d", m.Chat.ID, m.ReplyTo.Sender.ID)
				if _, ok := zcomap.Get(token); ok {
					addCredit(m.Chat.ID, m.Sender, -10, true)
					SmartSend(m, "😠 你自己先漱漱口呢，不要连续臭别人哦！扣 10 分警告一下")
				} else if senderLimit, _ := zcomap.Get(limSenderToken); senderLimit >= 2 {
					zcomap.Add(limReciverToken)
					SmartSend(m, "😳 用指令对线是不对的，请大家都冷静下呢～")
				} else {
					zcomap.Add(limSenderToken)
					zcomap.Add(limReciverToken)
					zcomap.Set(token, 1)
					ci := addCredit(m.Chat.ID, m.ReplyTo.Sender, -25, true)
					SmartSend(m.ReplyTo, fmt.Sprintf("%s, 您被热心的 %s 警告了 ⚠️，请注意管理好自己的行为！暂时扣除 25 分作为警告，如果您的分数低于 -50 分将被直接禁言。若您觉得这是恶意举报，请理性对待，并联系群管理员处理。", GetUserName(m.ReplyTo.Sender), GetUserName(m.Sender)))
					LazyDelete(m)
					if ci.Credit < -50 {
						Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0)
					}
				}
			}
		} else {
			SmartSend(m, "😠 匿名就不要乱啵啵啦！叭了个叭叭了个叭叭了个叭 ...")
		}
	}
}

func CMDBanUser(m *tb.Message) {
	if IsGroup(m.Chat.ID) && m.ReplyTo != nil {
		if m.Sender.ID > 0 && m.Sender.Username != "Channel_Bot" {
			if m.ReplyTo.Sender.ID == m.Sender.ID {
				if Ban(m.Chat.ID, m.Sender.ID, 1800) == nil {
					SmartSend(m, "举报自己？那没办法...只好把你 🫒 半小时哦～")
					LazyDelete(m.ReplyTo)
				} else {
					SmartSend(m, "呜呜呜，封不掉 ～")
				}
			} else if m.ReplyTo.Sender.IsBot && m.ReplyTo.SenderChat != nil {
				if m.ReplyTo.SenderChat != nil && m.ReplyTo.SenderChat.ID != m.Chat.ID {
					if BanChannel(m.Chat.ID, m.ReplyTo.SenderChat.ID) == nil {
						SmartSend(m, fmt.Sprintf("好的！这就把这个频道封掉啦～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）", GetChatName(m.ReplyTo.SenderChat)))
						LazyDelete(m)
						LazyDelete(m.ReplyTo)
					} else {
						SmartSend(m, "呜呜呜，封不掉 ～")
					}
				} else {
					SmartSend(m, "叭了个叭叭了个叭叭了个叭 ～")
				}
			} else if m.ReplyTo.Sender.IsBot {
				if Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 1800) == nil {
					SmartSend(m, fmt.Sprintf("好的！这就把这个机器人封禁半小时～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）", GetUserName(m.ReplyTo.Sender)))
					LazyDelete(m)
					LazyDelete(m.ReplyTo)
				} else {
					SmartSend(m, "呜呜呜，封不掉 ～")
				}
			} else {
				userId := m.ReplyTo.Sender.ID
				vtToken := fmt.Sprintf("vt-%d,%d", m.Chat.ID, userId)
				token := fmt.Sprintf("ad-%d,%d", m.Chat.ID, m.Sender.ID)
				if zcomap.Add(token) > 3 {
					addCredit(m.Chat.ID, m.Sender, -5, true)
					SmartSend(m, "😠 消停一下消停一下，举报太多次啦，扣 5 分缓一缓")
				} else {
					if _, ok := votemap.Get(vtToken); !ok {
						if Ban(m.Chat.ID, userId, 1800) == nil {
							addCredit(m.Chat.ID, m.ReplyTo.Sender, -50, true)
							addCredit(m.Chat.ID, m.Sender, 15, true)
							votemap.Set(vtToken, 0)
							msgTxt := fmt.Sprintf("%s, 您被热心群友 %s 报告有发送恶意言论的嫌疑 ⚠️，请注意自己的发言哦！暂时禁言半小时并扣除 50 分作为警告，举报者 15 分奖励已到账。若您觉得这是恶意举报，可以呼吁小伙伴们公投为您解封（累计满 6 票可以解封并抵消扣分），或者直接联系群管理员处理。", GetUserName(m.ReplyTo.Sender), GetUserName(m.Sender))
							SendBtns(m.ReplyTo, msgTxt, "", GenVMBtns(0, m.Chat.ID, userId, m.Sender.ID))
							LazyDelete(m)
							LazyDelete(m.ReplyTo)
						} else {
							SmartSend(m, "呜呜呜，封不掉 ～")
						}
					} else {
						SmartSend(m, "他已经被检察官带走啦，不要鞭尸啦 ～")
					}
				}
			}
		} else {
			SmartSend(m, "😠 匿名就不要乱啵啵啦！叭了个叭叭了个叭叭了个叭 ...")
		}
	}
}
