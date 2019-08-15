package client

import (
	"bytes"
	"log"
	"regexp"
	"strings"
)

var link = regexp.MustCompile(`\[\[(.*?)(?:\||]])`)
var template = regexp.MustCompile(`{{((?:.)*?)[\n|]`)
var parameter = regexp.MustCompile(`\|\s*?(.*?)\s`)
var nowiki = regexp.MustCompile(`<nowiki>((?:.|\n)*?)<\/nowiki>`)

func (w *WikiClient) CompareTranslations(title string) {
	titles := []string{title}
	api, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[CompareTranslations] Error getting articles:\n%s", err.Error())
	}

	links := w.GetLinks(api[title])
	log.Println(links)
}

func (w *WikiClient) GetLinks(article []byte) map[string]int {
	links := link.FindAllSubmatchIndex(article, -1)

	linkDict := make(map[string]int)

	for _, link := range links {
		linkString := string(article[link[2]:link[3]])
		count, exists := linkDict[linkString]

		if exists {
			linkDict[linkString] = count + 1
		} else {
			linkDict[linkString] = 1
		}
	}
	return linkDict
}

func (w *WikiClient) GetRedirects(titles []string) []string {
	articles, err := w.GetArticles(titles)
	if err != nil {
		log.Printf("[GetRedirects] Error: %s", err)
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

func (w *WikiClient) IsLink(template string) bool {
	lower := strings.ToLower(template)
	return (lower == "item link" || lower == "update link" || lower == "class link")
}
