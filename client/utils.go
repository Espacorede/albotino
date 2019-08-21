package client

import (
	"regexp"
	"strings"
)

var ignoreLinkRegexp = regexp.MustCompile(`(?i)^(w(ikia|ikipedia)?|p2|vdc):`)

var firstCharRegexp = regexp.MustCompile(`^[a-z]`)

func isIgnoreLink(link string) bool {
	return ignoreLinkRegexp.MatchString(link)
}

func IsIgnoreTemplate(template string) bool {
	lower := strings.ToLower(template)
	return strings.HasSuffix(lower, "link") ||
		strings.HasSuffix(lower, "name") ||
		lower == "common string" ||
		lower == "trans"
}

func Title(str string) string {
	return firstCharRegexp.ReplaceAllStringFunc(str, func(match string) string {
		return strings.ToUpper(match)
	})
}

func sumMap(m map[string]int) int {
	diff := 0
	for _, value := range m {
		if value < 0 {
			diff -= value
		} else {
			diff += value
		}
	}
	return diff
}

// nota pra mim mesmo: o parâmetro "linksLang" é só pra calcular os mapas de LINKS
func mapDifference(english map[string]int, translation map[string]int, linksLang string) map[string]int {
	difference := make(map[string]int)
	for key, value := range english {
		var transLinks int
		if linksLang != "" {
			langLink := key + "/" + linksLang
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

func GetTemplates(article string) map[string]int {
	templates := template.FindAllStringSubmatch(article, -1)
	templatesSlice := []string{}
	for _, template := range templates {
		templateName := Title(template[1])
		if IsIgnoreTemplate(templateName) || strings.HasPrefix(templateName, "DISPLAYTITLE:") {
			continue
		}
		templatesSlice = append(templatesSlice, templateName)
	}

	templatesDict := make(map[string]int)

	for _, templateName := range templatesSlice {
		templatesDict[templateName]++
	}
	return templatesDict
}

func GetParameters(article string) map[string]int {
	parameters := parameter.FindAllStringSubmatch(article, -1)
	parametersDict := make(map[string]int)

	for _, param := range parameters {
		parameterName := param[1]
		parametersDict[parameterName]++
	}
	return parametersDict
}

func GetHeaders(article string) int {
	headers := header.FindAllString(article, -1)
	return len(headers)
}
