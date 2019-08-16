package client

import (
	"bytes"
	"log"
	"regexp"
	"strings"
)

var languages = []string{"ar", "cs", "da", "de", "es", "fi", "fr", "hu", "it", "ja", "ko", "nl", "no", "pl", "pt", "pt-br", "ro", "ru", "sv", "tr", "zh-hans", "zh-hant"}

var link = regexp.MustCompile(`\[\[(.+?)(?:\||]])`)
var template = regexp.MustCompile(`{{((?:.)+?)(?:\n|}}|\|)`)
var parameter = regexp.MustCompile(`\|\s*?(.+?)\s`)
var templateLinks = regexp.MustCompile(`(?i){{(?:update|item|class) link\|(.+?)(?:}}|\|)`)

func (w *WikiClient) CompareTranslations(title string) {
	var trimTitle string
	nonEnglish := strings.LastIndex(title, "/")
	if nonEnglish == -1 {
		trimTitle = title
	} else {
		trimTitle = title[0:nonEnglish]
	}
	titles := []string{trimTitle}

	for _, lang := range languages {
		titles = append(titles, trimTitle+"/"+lang)
	}

	api, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[CompareTranslations] Error getting articles->\n\t%s", err.Error())
	}

	links := w.GetLinks(api[trimTitle], "")
	templates := GetTemplates(api[trimTitle])
	log.Println(links)
	log.Println(templates)
	for key, value := range api {
		if key == trimTitle || value == nil {
			continue
		}
		lang := title[(strings.LastIndex(title, "/") + 1):len(title)]
		log.Println(w.GetLinks(value, lang))
		log.Println(GetTemplates(value))
	}
}

func (w *WikiClient) GetLinks(article []byte, lang string) map[string]int {
	links := link.FindAllSubmatch(article, -1)
	fromTemplates := templateLinks.FindAllSubmatch(article, -1)
	linkSlice := []string{}

	for _, link := range links {
		linkSlice = append(linkSlice, string(link[1]))
	}
	for _, link := range fromTemplates {
		linkSlice = append(linkSlice, string(link[1])+lang)
	}

	linkSlice = w.GetRedirects(linkSlice)

	linkDict := make(map[string]int)

	for _, linkString := range linkSlice {
		if isIgnoreLink(linkString) {
			continue
		}
		count := linkDict[linkString]

		linkDict[linkString] = count + 1
	}
	return linkDict
}

func (w *WikiClient) GetRedirects(titles []string) []string {
	articles, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[GetRedirects] Error->\n\t%s", err)
		return nil
	}
	redirectTitles := make([]string, len(titles))
	for index, name := range titles {
		article := articles[name]
		if bytes.Index(article, []byte("# REDIRECT")) == 0 {
			redirectTitles[index] = string(link.Find([]byte(article)))
		} else {
			redirectTitles[index] = name
		}
	}

	return redirectTitles
}
