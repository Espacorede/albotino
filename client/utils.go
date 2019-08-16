package client

import (
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
		count := templatesDict[templateName]

		templatesDict[templateName] = count + 1
	}
	return templatesDict
}
