package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdOnCallback(c *tb.Callback) {
	m := c.Message
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil {
		callbacklock.Lock()
		defer callbacklock.Unlock()

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
		triggerUid := c.Sender.ID
		vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)
		joinVerificationId := fmt.Sprintf("join,%d,%d", gid, uid)
		isGroupAdmin := IsGroupAdmin(m.Chat, c.Sender)
		isMiaoGroupAdmin := IsGroupAdminMiaoKo(m.Chat, c.Sender)
		if strings.Contains("vt unban kick check rp lt", cmd) && IsGroup(gid) && uid > 0 {
			if cmd == "unban" && isGroupAdmin {
				if Unban(gid, uid, 0) == nil {
					Rsp(c, "✔️ 已解除封禁，请您手动处理后续事宜 ~")
				} else {
					Rsp(c, "❌ 解封失败，TA 可能已经被解封或者已经退群啦 ~")
				}
				SmartEdit(m, m.Text+"\n\nTA 已被管理员解封 👊")
				joinmap.Unset(joinVerificationId)
				if secuid > 0 && votemap.Exist(vtToken) {
					addCredit(gid, &tb.User{ID: uid}, 50, true)
					votemap.Unset(vtToken)
					addCredit(gid, &tb.User{ID: secuid}, -15, true)
				}
			} else if cmd == "kick" && isGroupAdmin {
				if Kick(gid, uid) == nil {
					Rsp(c, "✔️ 已将 TA 送出群留学去啦 ~")
				} else {
					Rsp(c, "❌ 踢出失败，可能 TA 已经退群啦 ~")
				}
				joinmap.Unset(joinVerificationId)
				votemap.Unset(vtToken)
				SmartEdit(m, m.Text+"\n\nTA 已被管理员踢出群聊 🦶")
			} else if cmd == "check" {
				if uid == c.Sender.ID {
					usrStatus := UserIsInGroup(gc.MustFollow, uid)
					if usrStatus == UIGIn {
						if Unban(gid, uid, 0) == nil {
							Bot.Delete(m)
							Rsp(c, "✔️ 验证成功，欢迎您的加入 ~")
							joinmap.Unset(joinVerificationId)
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
			} else if cmd == "rp" {
				redpacketKey := fmt.Sprintf("%d-%d", gid, secuid)

				credits, _ := redpacketmap.Get(redpacketKey)
				left, _ := redpacketnmap.Get(redpacketKey)
				if credits > 0 && left > 0 {
					redpacketBestKey := fmt.Sprintf("%d-%d:best", gid, secuid)
					redpacketUserKey := fmt.Sprintf("%d-%d:%d", gid, secuid, triggerUid)
					if redpacketmap.Add(redpacketUserKey) == 1 {
						amount := 0
						if left <= 1 {
							amount = credits
						} else if left == 2 {
							amount = rand.Intn(credits)
						} else {
							rate := 3
							if left <= 4 {
								rate = 2
							} else if left >= 12 {
								rate = 4
							}
							amount = rand.Intn(credits * rate / left)
						}
						redpacketnmap.Set(redpacketKey, left-1)
						redpacketmap.Set(redpacketKey, credits-amount)

						if amount == 0 {
							Rsp(c, "🐢 您的运气也太差啦！什么都没有抽到哦...")
						} else {
							lastBest, _ := redpacketmap.Get(redpacketBestKey)
							if amount > lastBest {
								redpacketmap.Set(redpacketBestKey, amount)
								redpacketrankmap.Set(redpacketBestKey, GetQuotableUserName(c.Sender))
							}
							Rsp(c, "🎉 恭喜获得 "+strconv.Itoa(amount)+" 积分，积分已经实时到账～")
							addCredit(gid, c.Sender, int64(amount), true)
						}

						SendRedPacket(m, gid, secuid)
					} else {
						Rsp(c, "❌ 您已经参与过这次活动了，不能太贪心哦！")
					}
				} else {
					Rsp(c, "❌ 抽奖活动已经结束啦！请期待下一次活动～")
				}
			} else if cmd == "lt" {
				cmdtype := uid // 做了转换 1: lottery, 2: start, 3: draw
				lotteryId := cmds[3]
				li := GetLottery(lotteryId)
				if li != nil {
					if cmdtype == 2 && isMiaoGroupAdmin {
						li.Status = 0
						li.Update()
						li.UpdateTelegramMsg()
						Rsp(c, "🎉 活动已确认，请号召群友踊跃参与哦！")
					} else if cmdtype == 3 && isMiaoGroupAdmin {
						li.CheckDraw(true)
					} else if cmdtype == 1 {
						ci := GetCredit(li.GroupID, triggerUid)
						if ci != nil {
							if ci.Credit >= int64(li.Limit) {
								if li.Consume {
									addCredit(li.GroupID, c.Sender, -int64(li.Limit), true)
								}
								if err := li.Join(triggerUid, GetQuotableUserName(c.Sender)); err == nil {
									Rsp(c, "🎉 参与成功 ~ 请耐心等待开奖呀 ~")
									if li.Participant > 0 {
										// check draw by particitant
										li.CheckDraw(false)
									}
									debouncer(func() {
										if li.Status == 0 {
											li.UpdateTelegramMsg()
										}
									})
								} else {
									if li.Consume {
										addCredit(li.GroupID, c.Sender, int64(li.Limit), true)
									}
									Rsp(c, err.Error())
								}
							} else {
								Rsp(c, "❌ 你的积分不满足活动要求哦！")
							}
						} else {
							Rsp(c, "❌ 请加群后再参与活动哦！")
						}
					} else {
						Rsp(c, "❌ 请不要乱玩喵组管理员指令！")
					}
				} else {
					Rsp(c, "❌ 未找到这个活动，请联系管理员解决！")
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
}
