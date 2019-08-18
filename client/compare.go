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
var parameter = regexp.MustCompile(`\| *(\w+?) *={1}`)
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

	english := api[trimTitle]

	links := w.GetLinks(english, "")
	templates := GetTemplates(english)
	parameters := GetParameters(english)
	for key, value := range api {
		if key == trimTitle || value == nil {
			continue
		}
		log.Println(key)
		lang := key[(strings.LastIndex(key, "/") + 1):len(key)]
		langLinks := w.GetLinks(value, lang)
		langTemplates := GetTemplates(value)
		langParameters := GetParameters(value)

		log.Println(mapDifference(links, langLinks, lang))
		log.Println(mapDifference(templates, langTemplates, ""))
		log.Println(mapDifference(parameters, langParameters, ""))
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
		linkSlice = append(linkSlice, string(link[1])+"/"+lang)
	}

	finalLinks := w.GetRedirects(linkSlice)

	linkDict := make(map[string]int)

	for _, linkString := range finalLinks {
		if isIgnoreLink(linkString) {
			continue
		}
		linkDict[linkString]++
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
		if bytes.Index(article, []byte("#REDIRECT")) == 0 {
			redirectTitles[index] = string(link.FindSubmatch(article)[1])
		} else {
			redirectTitles[index] = name
		}
	}

	return redirectTitles
}
