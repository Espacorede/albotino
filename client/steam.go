package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
)

const steamLocation string = `C:\Program Files (x86)\Steam\steamapps\common\Team Fortress 2\tf\resource`

var steamLocale map[string]string = map[string]string{
	"cs":      "tf_czech.txt",
	"da":      "tf_danish.txt",
	"de":      "tf_german.txt",
	"es":      "tf_spanish.txt",
	"fi":      "tf_finnish.txt",
	"fr":      "tf_french.txt",
	"hu":      "tf_hungarian.txt",
	"it":      "tf_italian.txt",
	"ja":      "tf_japanese.txt",
	"ko":      "tf_korean.txt",
	"nl":      "tf_dutch.txt",
	"no":      "tf_norwegian.txt",
	"pl":      "tf_polish.txt",
	"pt":      "tf_portuguese.txt",
	"pt-br":   "tf_brazilian.txt",
	"ro":      "tf_romanian.txt",
	"ru":      "tf_russian.txt",
	"sv":      "tf_swedish.txt",
	"tr":      "tf_turkish.txt",
	"zh-hans": "tf_schinese.txt",
	"zh-hant": "tf_tchinese.txt"}

func GetToken(itemName string) (string, error) {
	// podia ler do tf_english, mas prefiro ler de um traduzido porque dá uma leve especificidade maior pro regex com o [english].
	// o pt-br é bem completo, então o usaremos.
	localizationFile, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", steamLocation, steamLocale["pt-br"]))
	if err != nil {
		return "", err
	}

	itemTokenRegexp, err := regexp.Compile(fmt.Sprintf(`"\[english](.+?)"\t*"(The )?%s`, itemName))
	if err != nil {
		return "", err
	}
	tokenMatch := itemTokenRegexp.FindStringSubmatch(string(localizationFile))
	if tokenMatch == nil {
		return "", errors.New("Token not found.")
	}
	token := tokenMatch[1]
	return token, nil
}

func GetDescription(token string, lang string) (string, error) {
	file := fmt.Sprintf(`%s\%s`, steamLocation, steamLocale[lang])
	localizationFile, err := ioutil.ReadFile(file)

	descriptionRegexp, err := regexp.Compile(fmt.Sprintf(`"%s_desc"\t*"(.+?)"`, token))
	if err != nil {
		return "", err
	}
	descriptionMatch := descriptionRegexp.FindStringSubmatch(string(localizationFile))
	if descriptionMatch == nil {
		return "", errors.New("Description not found")
	}
	description := descriptionMatch[1]

	return description, nil
}
