package client

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
)

var languages = []string{"ar", "cs", "da", "de", "es", "fi", "fr", "hu", "it", "ja", "ko", "nl", "no", "pl", "pt", "pt-br", "ro", "ru", "sv", "tr", "zh-hans", "zh-hant"}

var link = regexp.MustCompile(`\[\[(.+?)(?:\||]])`)
var template = regexp.MustCompile(`{{((?:.)+?)(?:\n|}}|\|)`)
var parameter = regexp.MustCompile(`\| *([\w- ]+?) *=`)
var templateLinks = regexp.MustCompile(`(?i){{(?:update|item|class) link\|(.+?)(?:}}|\|)`)

var redirectRegexp = regexp.MustCompile(`(?i)#redirect \[\[(.*?)]]`)

var categoryRegexp = regexp.MustCompile(`(?i){{(scout|soldier|pyro|demoman|heavy|engineer|medic|sniper|spy|hat|allweapons) nav}}`)

func (w *WikiClient) ProcessArticle(title string) {
	var trimTitle string
	nonEnglish := strings.LastIndex(title, "/")
	if nonEnglish == -1 {
		trimTitle = strings.Trim(title, " ")
	} else {
		trimTitle = strings.Trim(title[0:nonEnglish], " ")
	}
	titles := []string{trimTitle}
	api, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[CompareTranslations] Error getting articles->\n\t%s\n", err.Error())
		return
	}

	englishPage := api[trimTitle]

	if englishPage.article == "" {
		log.Printf("Page %s not found.", trimTitle)
	} else if englishPage.namespace != 0 {
		log.Printf("%s is not main; ignoring", trimTitle)
		return
	} else {
		log.Printf("Comparing translations for %s", trimTitle)
		w.CompareTranslations(trimTitle, englishPage.article)
	}
}

func (w *WikiClient) CompareTranslations(title string, english string) {
	titles := []string{}

	for _, lang := range languages {
		titles = append(titles, title+"/"+lang)
	}

	api, err := w.GetArticles(titles)

	if err != nil {
		log.Printf("[CompareTranslations] Error getting articles->\n\t%s\n", err.Error())
		return
	}

	englishBytes := len([]byte(english))

	links, _ := w.GetLinks(english, "")
	templates := GetTemplates(english)
	parameters := GetParameters(english)

	categories := categoryRegexp.FindAllStringSubmatch(english, -1)

	var class string
	var itemtype string
	var category string

	for _, match := range categories {
		switch strings.ToLower(match[1]) {
		case "hat":
			itemtype = "cosmetics"
			break
		case "allweapons":
			itemtype = "weapons"
			break
		default:
			if class == "" {
				class = match[1]
			} else if class != "multiclass" {
				class = "multiclass"
			}
		}
	}

	if itemtype == "" {
		category = "others"
	} else if class == "" {
		if itemtype == "cosmetics" {
			category = "allclass cosmetics"
		} else {
			category = "weapons"
		}
	} else if itemtype == "cosmetics" {
		category = fmt.Sprintf("%s cosmetics", class)
	} else {
		category = "weapons"
	}

	englishPoints := sumMap(links)*2 + sumMap(templates)*3 + sumMap(parameters)

	languageValues := make([]int64, len(languages))

	for i := range languages {
		languageValues[i] = -1
	}

	for key, value := range api {
		if value.article == "" {
			continue
		}

		lang := key[(strings.LastIndex(key, "/") + 1):len(key)]

		langindex := -1

		for index, icon := range languages {
			if icon == lang {
				langindex = index
				break
			}
		}

		langPage := value.article

		langLinks, wrongLinks := w.GetLinks(langPage, lang)

		langTemplates := GetTemplates(langPage)
		langParameters := GetParameters(langPage)

		linkDiff := mapDifference(links, langLinks, lang)
		templateDiff := mapDifference(templates, langTemplates, "")
		parametersDiff := mapDifference(parameters, langParameters, "")

		linkPoints := sumMap(linkDiff)
		templatePoints := sumMap(templateDiff)
		parameterPoints := sumMap(parametersDiff)

		languagePoints := linkPoints*2 + templatePoints*3 + parameterPoints + len(wrongLinks)*2

		updatePoints := int64(math.Round(float64(languagePoints) / float64(englishPoints) * float64(englishBytes)))
		languageValues[langindex] = updatePoints

		//log.Printf("%s: %d points", key, updatePoints)
	}

	err = upsertDBEntry(title, languageValues, category)
	if err != nil {
		log.Printf("[CompareTranslations] Error inserting %s values to db->\n\t%s\n", title, err)
	}
}

func (w *WikiClient) GetLinks(article string, lang string) (map[string]int, []string) {
	links := link.FindAllStringSubmatch(article, -1)
	fromTemplates := templateLinks.FindAllStringSubmatch(article, -1)
	linkSlice := []string{}
	wrongLanguage := []string{}

	for _, link := range links {
		title := Title(link[1])
		linkSlice = append(linkSlice, title)

		if !strings.HasSuffix(title, lang) {
			wrongLanguage = append(wrongLanguage, title)
		}
	}
	for _, link := range fromTemplates {
		linkSlice = append(linkSlice, link[1]+"/"+lang)
	}

	finalLinks := w.GetRedirects(linkSlice)

	linkDict := make(map[string]int)

	for _, linkString := range finalLinks {
		if isIgnoreLink(linkString) {
			continue
		}
		linkDict[linkString]++
	}
	return linkDict, wrongLanguage
}

func (w *WikiClient) GetRedirects(titles []string) []string {
	articles, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[GetRedirects] Error->\n\t%s\n", err)
		return nil
	}
	redirectTitles := make([]string, len(titles))
	for index, name := range titles {
		article := articles[name].article
		redirect := redirectRegexp.FindStringSubmatch(article)
		if redirect == nil {
			redirectTitles[index] = name
		} else {
			redirectTitles[index] = redirect[1]

		}
	}
	return redirectTitles
}
