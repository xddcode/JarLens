package ui

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"

	jarlensupdate "github.com/xddcode/jarlens/internal/update"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

func TestInterfaceTranslationCoverage(t *testing.T) {
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel == nil || selector.Sel.Name != "t" && selector.Sel.Name != "tf" || len(call.Args) == 0 {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			source, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Errorf("%s: invalid translated literal %s", file.Name(), literal.Value)
				return true
			}
			for _, language := range []string{languageTraditionalChinese, languageEnglish, languageJapanese, languageKorean} {
				if interfaceTranslations[language][source] == "" {
					t.Errorf("%s: %q has no %s translation", file.Name(), source, language)
				}
			}
			return true
		})
	}
}

func TestSettingsTranslationCoverage(t *testing.T) {
	m := &Model{}
	for _, language := range languageIDs {
		m.settings.Language = language
		for _, page := range settingsPages {
			texts := []string{page.title, page.description}
			texts = append(texts, page.labels...)
			for _, source := range texts {
				if source == "" || !strings.ContainsFunc(source, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
					continue
				}
				if language != languageChinese && m.t(source) == source {
					t.Errorf("%s: %q is not translated", language, source)
				}
			}
		}
	}
}

func TestSettingsLanguageSwitchAndSearch(t *testing.T) {
	tests := []struct {
		language       string
		option         int
		wantSettings   string
		wantAppearance string
		wantFont       string
		wantQuery      string
	}{
		{language: languageTraditionalChinese, option: 1, wantSettings: "設定", wantAppearance: "外觀", wantFont: "字型", wantQuery: "程式碼字型"},
		{language: languageEnglish, option: 2, wantSettings: "Settings", wantAppearance: "Appearance", wantFont: "Font", wantQuery: "Code Font"},
		{language: languageJapanese, option: 3, wantSettings: "設定", wantAppearance: "外観", wantFont: "フォント", wantQuery: "コードフォント"},
		{language: languageKorean, option: 4, wantSettings: "설정", wantAppearance: "모양", wantFont: "글꼴", wantQuery: "코드 글꼴"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
			m := &Model{SettingsStore: store}
			driver := newDriver(t, m)
			defer driver.Close()
			openSettingsForTest(t, driver)
			click(t, driver, settingsLanguage.At(5))
			click(t, driver, settingsLanguage.At(test.option))
			click(t, driver, settingsApply)
			if store.saves != 1 || m.settings.Language != test.language {
				t.Fatalf("language was not applied and persisted: %+v", m.settings)
			}
			if !shows(driver, test.wantSettings) || !hasHit(driver, settingsNav.At(1), test.wantAppearance) {
				t.Fatalf("settings page did not switch to %s", test.language)
			}
			click(t, driver, settingsSearch)
			if err := driver.Type(test.wantQuery); err != nil {
				t.Fatal(err)
			}
			if m.settingsCategory != 3 || !hasHit(driver, settingsNav.At(3), test.wantFont) {
				t.Fatalf("localized settings search did not find %q", test.wantFont)
			}
		})
	}
}

func TestLocalizedUpdateStatus(t *testing.T) {
	tests := []struct {
		language string
		want     string
	}{
		{language: languageTraditionalChinese, want: "目前已是最新版本"},
		{language: languageEnglish, want: "You are already using the latest version"},
		{language: languageJapanese, want: "すでに最新バージョンです"},
		{language: languageKorean, want: "이미 최신 버전을 사용 중입니다"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			settings := jarlensworkspace.DefaultSettings()
			settings.Language = test.language
			m := &Model{
				SettingsStore: &memorySettingsStore{settings: settings},
				UpdateChecker: func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error) {
					return jarlensupdate.Release{Version: jarlensupdate.Current}, nil
				},
			}
			driver := newDriver(t, m)
			defer driver.Close()
			m.CheckForUpdates(false)
			pumpUntil(t, driver, func() bool { return m.updateStatus == test.want })
		})
	}
}

func TestTranslationCatalogsCoverTheSameMessages(t *testing.T) {
	english := interfaceTranslations[languageEnglish]
	for _, language := range []string{languageTraditionalChinese, languageJapanese, languageKorean} {
		catalog := interfaceTranslations[language]
		for source := range english {
			if catalog[source] == "" {
				t.Errorf("%s: missing translation for %q", language, source)
			}
		}
		for source := range catalog {
			if english[source] == "" {
				t.Errorf("%s: obsolete translation for %q", language, source)
			}
		}
	}
}
