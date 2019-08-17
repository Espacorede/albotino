package client

import (
	"log"
	"regexp"
	"strings"
)

var ignoreLinkRegexp = regexp.MustCompile(`(?i)^(w(ikia|ikipedia)?|p2|vdc):`)

func isIgnoreLink(link string) bool {
	return ignoreLinkRegexp.MatchString(link)
}

func IsLink(template string) bool {
	return strings.HasSuffix(strings.ToLower(template), "link")
}

func mapDifference(english map[string]int, translation map[string]int, lang string) map[string]int {
	difference := make(map[string]int)
	for key, value := range english {
		var transLinks int
		if lang != "" {
			langLink := key + "/" + lang
			transLinks += translation[langLink]
		}
		transLinks += translation[key]
		difference[key] = value - transLinks
	}
	for key, value := range translation {
		keySeparator := strings.LastIndex(key, "/")
		var link string
		if keySeparator == -1 {
			link = key
		} else {
			link = key[0:keySeparator]
		}
		enCount, enExists := english[link]
		if enExists {
			continue
		}
		difference[link] = enCount - value
	}
	return difference
}

func GetTemplates(article []byte) map[string]int {
	templates := template.FindAllSubmatch(article, -1)
	templatesSlice := []string{}
	for _, template := range templates {
		templateName := string(template[1])
		if IsLink(templateName) || strings.HasPrefix(templateName, "DISPLAYTITLE:") {
			continue
		}
		templatesSlice = append(templatesSlice, templateName)
	}

	templatesDict := make(map[string]int)

	for _, templateName := range templatesSlice {
		if templateName == "Item infobox\n" {
			log.Println(templates)
		}
		templatesDict[templateName]++
	}
	return templatesDict
}

func GetParameters(article []byte) map[string]int {
	parameters := parameter.FindAllSubmatch(article, -1)
	parametersDict := make(map[string]int)

	for _, param := range parameters {
		parameterName := string(param[1])
		parametersDict[parameterName]++
	}
	return parametersDict
}
