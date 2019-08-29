package client

import (
	"log"
	"math"
	"regexp"
	"strings"
)

var languages = []string{"ar", "cs", "da", "de", "es", "fi", "fr", "hu", "it", "ja", "ko", "nl", "no", "pl", "pt", "pt-br", "ro", "ru", "sv", "tr", "zh-hans", "zh-hant"}

var link = regexp.MustCompile(`\[\[(.+?)(?:\||]])`)
var template = regexp.MustCompile(`{{((?:.)+?)(?:\n|}}|\|)`)
var parameter = regexp.MustCompile(`\| *(\w+?) *={1}`)
var templateLinks = regexp.MustCompile(`(?i){{(?:update|item|class) link\|(.+?)(?:}}|\|)`)

var redirectRegexp = regexp.MustCompile(`(?i)#redirect \[\[(.*?)]]`)

func (w *WikiClient) CompareLinks(article string) error {
	linkMatches := link.FindAllStringSubmatch(article, -1)

	links := []string{}

	for _, link := range linkMatches {
		links = append(links, link[1])
	}

	w.CompareMultiple(links)

	return nil
}

func (w *WikiClient) CompareMultiple(titles []string) {
	for _, article := range titles {
		w.ProcessArticle(article, false)
	}
}

func (w *WikiClient) ProcessArticle(title string, recursion bool) {
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
		log.Printf("[CompareTranslations] Error getting articles->\n\t%s", err.Error())
		return
	}

	englishPage := api[trimTitle]

	if englishPage.article == "" {
		log.Println("Page " + trimTitle + " not found.")
	} else if englishPage.namespace != 0 {
		if recursion {
			log.Println("Comparing links for " + trimTitle)
			w.CompareLinks(englishPage.article)
		} else {
			log.Println(trimTitle + " is not main; ignoring")
			return
		}
	} else {
		log.Println("Comparing translations for " + trimTitle)
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
		log.Printf("[CompareTranslations] Error getting articles->\n\t%s", err.Error())
		return
	}

	englishBytes := len([]byte(english))

	links, _ := w.GetLinks(english, "")
	templates := GetTemplates(english)
	parameters := GetParameters(english)

	englishPoints := float64(sumMap(links) + sumMap(templates) + sumMap(parameters))

	for key, value := range api {
		if value.article == "" {
			continue
		}

		log.Println(key)
		lang := key[(strings.LastIndex(key, "/") + 1):len(key)]

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

		languagePoints := float64(linkPoints + templatePoints + parameterPoints + len(wrongLinks))

		updatePoints := math.Round((languagePoints / englishPoints) * float64(englishBytes))

		log.Println(updatePoints)
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
		log.Printf("[GetRedirects] Error->\n\t%s", err)
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
