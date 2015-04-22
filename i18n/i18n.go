package i18n

import (
	"io/ioutil"

	"github.com/ije/gox/config"
)

var IETFLangs = map[string]string{
	"id":     "Bahasa Indonesia",
	"ms":     "Bahasa Melayu",
	"ca":     "Català",
	"cs":     "Čeština",
	"cy":     "Cymraeg",
	"da":     "Dansk",
	"de":     "Deutsch",
	"et":     "Eesti keel",
	"en":     "English",
	"en-GB":  "English (UK)",
	"en-US":  "English (US)",
	"es":     "Español",
	"es-419": "Español (Latinoamérica)",
	"eu":     "Euskara",
	"tl":     "Filipino",
	"fr":     "Français",
	"hr":     "Hrvatski",
	"it":     "Italiano",
	"is":     "Íslenska",
	"sw":     "Kiswahili",
	"lv":     "Latviešu",
	"lt":     "Lietuvių",
	"hu":     "Magyar",
	"no":     "Norsk (Bokmål)",
	"nl":     "Nederlands",
	"pl":     "Polski",
	"pt-BR":  "Português (Brasil)",
	"pt-PT":  "Português (Portugal)",
	"ro":     "Română",
	"sk":     "Slovenčina",
	"sl":     "Slovenščina",
	"fi":     "Suomi",
	"sv":     "Svenska",
	"vi":     "Tiếng Việt",
	"tr":     "Türkçe",
	"el":     "Ελληνικά",
	"bg":     "Български",
	"ru":     "Русский",
	"sr":     "Српски",
	"uk":     "Українська",
	"iw":     "‫עברית‬‎",
	"ur":     "‫اردو‬‎",
	"ar":     "‫العربية‬‎",
	"fa":     "‫فارسی‬‎",
	"mr":     "मराठी",
	"hi":     "हिन्दी",
	"bn":     "বাংলা",
	"gu":     "ગુજરાતી",
	"ta":     "தமிழ்",
	"te":     "తెలుగు",
	"kn":     "ಕನ್ನಡ",
	"ml":     "മലയാളം",
	"th":     "ภาษาไทย",
	"am":     "አማርኛ (Amharic)",
	"chr":    "ᏣᎳᎩ (Cherokee)",
	"zh-CN":  "中文（简体）",
	"zh-HK":  "中文（繁體）",
	"zh-TW":  "中文（繁體）",
	"ja":     "日本語",
	"ko":     "한국어",
}

type I18n struct {
	dict map[string]config.Section
}

func New(i18nFile, defaultLang string) (i18n *I18n, err error) {
	data, err := ioutil.ReadFile(i18nFile)
	if err != nil {
		return
	}
	defaultSection, extendedSections := config.Parse(data)
	if _, ok := extendedSections[defaultLang]; !ok {
		extendedSections[defaultLang] = defaultSection
	}
	i18n = &I18n{extendedSections}
	return
}

func (i18n *I18n) Translate(lang, text string) string {
	if dict, ok := i18n.dict[lang]; ok {
		if ret, ok := dict[text]; ok && len(ret) > 0 {
			return ret
		}
	}
	return text
}
