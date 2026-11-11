package ui

import "fmt"

const (
	languageChinese            = "zh-CN"
	languageTraditionalChinese = "zh-TW"
	languageEnglish            = "en"
	languageJapanese           = "ja"
	languageKorean             = "ko"
)

var (
	languageIDs   = []string{languageChinese, languageTraditionalChinese, languageEnglish, languageJapanese, languageKorean}
	languageNames = []string{"简体中文", "繁體中文", "English", "日本語", "한국어"}
)

func languageIndex(language string) int {
	for index, id := range languageIDs {
		if id == language {
			return index
		}
	}
	return 0
}

func normalizeLanguage(language string) string {
	if index := languageIndex(language); index >= 0 && index < len(languageIDs) && languageIDs[index] == language {
		return language
	}
	return languageChinese
}

func (m *Model) t(source string) string {
	language := normalizeLanguage(m.settings.Language)
	if language == languageChinese {
		return source
	}
	if translated := interfaceTranslations[language][source]; translated != "" {
		return translated
	}
	return source
}

func (m *Model) tf(source string, args ...any) string {
	return fmt.Sprintf(m.t(source), args...)
}
