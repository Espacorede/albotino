package client

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strings"
)

var languages = []string{"ar", "cs", "da", "de", "es", "fi", "fr", "hu", "it", "ja", "ko", "nl", "no", "pl", "pt", "pt-br", "ro", "ru", "sv", "tr", "zh-hans", "zh-hant"}

var link = regexp.MustCompile(`\[\[(.+?)(?:\||]])`)
var template = regexp.MustCompile(`{{((?:.)+?)(?:\n|}}|\|)`)
var parameter = regexp.MustCompile(`\| *(\w+?) *={1}`)
var templateLinks = regexp.MustCompile(`(?i){{(?:update|item|class) link\|(.+?)(?:}}|\|)`)

var descriptionRegexp = regexp.MustCompile(`\| item-description\s+= (.+)`)

var redirectRegexp = regexp.MustCompile(`(?i)#redirect \[\[(.*?)]]`)

func (w *WikiClient) ProcessArticle(title string, compareDescriptions bool) {
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
		log.Println(trimTitle + " is not main; ignoring")
		return
	} else {
		log.Println("Comparing translations for " + trimTitle)
		w.CompareTranslations(trimTitle, englishPage.article, compareDescriptions)
	}
}

func (w *WikiClient) CompareTranslations(title string, english string, compareDescriptions bool) {
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

	hasDescription := descriptionRegexp.MatchString(english)

	languageValues := make([]int64, len(languages))

	for i := range languages {
		languageValues[i] = -1
	}

	var descriptionFile *os.File
	var descriptionBuf *bufio.Writer
	var stsToken string

	if compareDescriptions && hasDescription {
		descriptionFile, err = os.OpenFile("temp/descriptions.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("[CompareTranslations] Error creating descriptions.txt->\n\t%s", err.Error())
		} else {
			defer descriptionFile.Close()
			descriptionBuf = bufio.NewWriter(descriptionFile)
		}

		stsToken, err = GetToken(title)
		if err != nil {
			log.Printf("[CompareTranslations] Error getting local token for %s->\n\t%s", title, err.Error())
		}
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

		languagePoints := float64(linkPoints + templatePoints + parameterPoints + len(wrongLinks))

		updatePoints := math.Round((languagePoints / englishPoints) * float64(englishBytes))

		languageValues[langindex] = int64(updatePoints)

		if descriptionFile != nil {
			stsDescription, err := GetDescription(stsToken, lang)
			if err != nil {
				log.Printf("[CompareTranslations] Error getting description for %s->\n\t%s", key, err.Error())
			}
			if parametersDiff["item-description"] <= 0 {
				descriptionMatch := descriptionRegexp.FindAllStringSubmatch(langPage, -1)
				if descriptionMatch == nil {
					log.Printf("[CompareTranslations] Error getting description from item-description parameter in %s. This really shouldn't happen.", key)
				}
				for _, match := range descriptionMatch {
					pageDescription := match[1]
					if stsDescription != pageDescription {
						descriptionBuf.WriteString(fmt.Sprintf("%s has a wrong description parameter. The official description is\n\n\t'%s'\n\n", key, stsDescription))
					}
				}
			} else {
				descriptionBuf.WriteString(fmt.Sprintf("%s is missing item-description parameter. Its official description is\n\n\t'%s'\n\n", key, stsDescription))
			}

			descriptionBuf.Flush()
		}

		log.Printf("%s: %f points", key, updatePoints)
	}

	upsertDBEntry(title, languageValues)
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
