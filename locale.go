package main

var LocaleMap = map[string]map[string]string{
	"default": {
		"cmd.zc.notAllowed":  "当前群组不允许互相臭嘴哦 ~",
		"cmd.zc.indeed":      "确实",
		"cmd.zc.cantBan":     "我拿它没办法呢 ...",
		"cmd.zc.cooldown10":  "😠 你自己先漱漱口呢，不要连续臭别人哦！扣 10 分警告一下",
		"cmd.zc.cooldown":    "😳 用指令对线是不对的，请大家都冷静下呢～",
		"cmd.zc.exec":        "%s, 您被热心的 %s 警告了 ⚠️，请注意管理好自己的行为！暂时扣除 25 分作为警告，如果您的分数低于 -50 分将被直接禁言。若您觉得这是恶意举报，请理性对待，并联系群管理员处理。",
		"cmd.zc.noAnonymous": "😠 匿名就不要乱啵啵啦！叭了个叭叭了个叭叭了个叭 ...",

		"cmd.ey.selfReport":  "举报自己？那没办法...只好把你 🫒 半小时哦～",
		"cmd.ey.notSuccess":  "呜呜呜，封不掉 ～",
		"cmd.ey.unexpected":  "叭了个叭叭了个叭叭了个叭 ～",
		"cmd.ey.killChannel": "好的！这就把这个频道封掉啦～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）",
		"cmd.ey.killBot":     "好的！这就把这个机器人封禁半小时～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）",
		"cmd.ey.cooldown5":   "😠 消停一下消停一下，举报太多次啦，扣 5 分缓一缓",
		"cmd.ey.exec":        "%s, 您被热心群友 %s 报告有发送恶意言论的嫌疑 ⚠️，请注意自己的发言哦！暂时禁言半小时并扣除 50 分作为警告，举报者 15 分奖励已到账。若您觉得这是恶意举报，可以呼吁小伙伴们公投为您解封（累计满 6 票可以解封并抵消扣分），或者直接联系群管理员处理。",
		"cmd.ey.duplicated":  "他已经被检察官带走啦，不要鞭尸啦 ～",

		"cmd.privateChatFirst": "❌ 请先私聊我然后再运行这个命令哦",
		"cmd.noGroupPerm":      "❌ 您没有权限，亦或是您未再对应群组使用这个命令",

		"credit.exportSuccess":    "\u200d 导出成功，请在私聊查看结果",
		"credit.importError":      "❌ 无法下载积分备份，请确定您上传的文件格式正确且小于 20MB，大文件请联系管理员手动导入",
		"credit.importParseError": "❌ 解析积分备份错误，请确定您上传的文件格式正确",

		"su.group.addSuccess":   "✔️ 已将该组加入积分统计 ～",
		"su.group.addDuplicate": "❌ 该组已经开启积分统计啦 ～",
		"su.group.delSuccess":   "✔️ 已将该组移除积分统计 ～",
		"su.group.delDuplicate": "❌ 该组尚未开启积分统计哦 ～",

		// not support yet
		"rp.complete":  "🧧 *积分红包*\n\n小伙伴们手速都太快啦，`%s`的大红包已被瓜分干净，没抢到的小伙伴们请期待下次的活动哦～",
		"rp.guessLeft": "猜猜看还剩多少？",
		"rp.text":      "🧧 *积分红包*\n\n``%s发红包啦！大家快抢哦～\n\n剩余积分: `%s`\n剩余数量: `%d`",
		"rp.lucky":     "\n\n恭喜手气王 `%s` 获得了 `%d` 分 🎉 ~",

		"channel.bot.permit":           "👏 欢迎 %s 加入群组，已为机器人自动放行 ～",
		"channel.user.alreadyFollowed": "👏 欢迎 %s 加入群组，您已关注频道自动放行 ～",
		"channel.request":              "[🎉](tg://user?id=%d) 欢迎 `%s`，您还没有关注本群组关联的频道哦，您有 5 分钟时间验证自己 ～ 请点击下面按钮跳转到频道关注后再回来验证以解除发言限制 ～",
		"channel.cannotSendMsg":        "❌ 无法发送验证消息，请管理员检查群组权限 ～",
		"channel.cannotBanUser":        "❌ 无法完成验证流程，请管理员检查机器人封禁权限 ～",
		"channel.cannotCheckChannel":   "❌ 无法检测用户是否在群组内，请管理员检查机器人权限 ～",
		"channel.kicked":               "👀 [TA](tg://user?id=%d) 没有在规定时间内完成验证，已经被我带走啦 ～",

		// not support yet
		"btn.rp.draw": "🤏 我要抢红包|rp/%d/1/%d",
		"btn.notFair": "😠 这不公平 (%d)|vt/%d/%d/%d",

		"btn.adminPanel":    "🚩 解封[管理]|unban/%d/%d/%d||🚮 清退[管理]|kick/%d/%d/%d",
		"btn.channel.step1": "👉 第一步：关注频道 👈|https://t.me/%s",
		"btn.channel.step2": "👉 第二步：点我验证 👈|check/%d/%d",

		"cb.unblock.byadmin": "\n\nTA 已被管理员解封 👊",
		"cb.kicked.byadmin":  "\n\nTA 已被管理员踢出群聊 🦶",
		"cb.unblock.byvote":  "\n\n于多名用户投票后决定，该用户不是恶意广告，用户已解封，积分已原路返回。",

		"cb.unban.success":                "✔️ 已解除封禁，请您手动处理后续事宜 ~",
		"cb.unban.failure":                "❌ 解封失败，TA 可能已经被解封或者已经退群啦 ~",
		"cb.kick.success":                 "✔️ 已将 TA 送出群留学去啦 ~",
		"cb.kick.failure":                 "❌ 踢出失败，可能 TA 已经退群啦 ~",
		"cb.validate.success":             "✔️ 验证成功，欢迎您的加入 ~",
		"cb.validate.success.cannotUnban": "❌ 验证成功，但是解禁失败，请联系管理员处理 ~",
		"cb.validate.failure":             "❌ 验证失败，请确认自己已经加入对应群组 ~",
		"cb.validate.others":              "😠 人家的验证不要乱点哦！！！",
		"cb.vote.success":                 "✔️ 投票成功，感谢您的参与 ~",
		"cb.vote.failure":                 "❌ 您已经参与过投票了，请不要多次投票哦 ~",
		"cb.vote.notExists":               "❌ 投票时间已过，请联系管理员处理 ~",
		"cb.rp.nothing":                   "🐢 您的运气也太差啦！什么都没有抽到哦...",
		"cb.rp.get.1":                     "🎉 恭喜获得 ",
		"cb.rp.get.2":                     " 积分，积分已经实时到账～",
		"cb.rp.duplicated":                "❌ 您已经参与过这次活动了，不能太贪心哦！",
		"cb.rp.notExists":                 "❌ 抽奖活动已经结束啦！请期待下一次活动～",
		"cb.lottery.start":                "🎉 活动已确认，请号召群友踊跃参与哦！",
		"cb.lottery.enroll":               "🎉 参与成功 ~ 请耐心等待开奖呀 ~",
		"cb.lottery.noEnoughCredit":       "❌ 你的积分不满足活动要求哦！",
		"cb.lottery.checkFailed":          "❌ 请加群后再参与活动哦！",
		"cb.notMiaoAdmin":                 "❌ 请不要乱玩喵组管理员指令！",
		"cb.notAdmin":                     "❌ 请不要乱玩管理员指令！",
		"cb.noEvent":                      "❌ 未找到这个活动，请联系管理员解决！",
		"cb.notParsed":                    "❌ 指令解析出错，请联系管理员解决 ~",
		"cb.disabled":                     "❌ 这个群组还没有被授权哦 ~",
	},
	"en": {},
}

func Locale(identifier string, locale string) string {
	if locales, ok := LocaleMap[locale]; ok && locales != nil {
		if text, ok := locales[identifier]; ok && text != "" {
			return text
		}
	}

	// fallback
	if text, ok := LocaleMap["default"][identifier]; ok && text != "" {
		return text
	}

	return identifier
}
