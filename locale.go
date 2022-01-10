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
