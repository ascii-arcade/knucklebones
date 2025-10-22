package language

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"
)

//go:embed en.json
var enJSON []byte

//go:embed es.json
var esJSON []byte

const (
	LanguageEN = "en"
	LanguageES = "es"
)

var Languages = map[string]*Language{
	LanguageEN: LoadLanguage(enJSON),
	LanguageES: LoadLanguage(esJSON),
}

type Language struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Translations map[string]any `json:"translations"`

	UsernameFirstWords  []string `json:"username_first_words"`
	UsernameSecondWords []string `json:"username_second_words"`
}

var DefaultLanguage = Languages["EN"]

func (l *Language) Get(pathList ...string) string {
	if len(pathList) == 0 {
		return ""
	}
	var current any = l.Translations
	for i, key := range pathList {
		m, ok := current.(map[string]any)
		if !ok {
			return missingTranslationValue(pathList)
		}
		val, exists := m[key]
		if !exists {
			return missingTranslationValue(pathList)
		}
		if i == len(pathList)-1 {
			strOut := ""
			if str, ok := val.(string); ok {
				strOut = str
			}

			strSlc, ok := val.([]interface{})
			if ok {
				for _, v := range strSlc {
					if str, ok := v.(string); ok {
						strOut += str + "\n"
					}
				}
			}

			if strOut == "" {
				return missingTranslationValue(pathList)
			}

			return strOut
		}
		current = val
	}
	return ""
}

func missingTranslationValue(pathList []string) string {
	return "i18n-missing:'" + strings.Join(pathList, ".") + "'"
}

func LoadLanguage(data []byte) *Language {
	var lang Language
	if err := json.Unmarshal(data, &lang); err != nil {
		log.Fatal("failed to decode language data: %w", err)
	}
	return &lang
}
