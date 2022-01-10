package main

Locale("", m.Sender.LanguageCode)

var LocaleMap = map[string]map[string]string{
	"default": {
		"cmd.privateChatFirst": "❌ 请先私聊我然后再运行这个命令哦",
		"cmd.noGroupPerm": "❌ 您没有权限，亦或是您未再对应群组使用这个命令",

		"credit.exportSuccess": "\u200d 导出成功，请在私聊查看结果",
		"credit.importError": "❌ 无法下载积分备份，请确定您上传的文件格式正确且小于 20MB，大文件请联系管理员手动导入",
		"credit.importParseError": "❌ 解析积分备份错误，请确定您上传的文件格式正确",

		"su.group.addSuccess": "✔️ 已将该组加入积分统计 ～",
		"su.group.addDuplicate": "❌ 该组已经开启积分统计啦 ～",
		"su.group.delSuccess": "✔️ 已将该组移除积分统计 ～",
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
