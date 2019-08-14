package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/buger/jsonparser"
)

const wiki string = "https://wiki.teamfortress.com/w/api.php"

// a fazer: estudar o que raios esses métodos precisam fazer
// https://golang.org/pkg/net/http/cookiejar/#PublicSuffixList
// talvez tenha a ver com isso? https://publicsuffix.org/
type suffixList struct {
}

func (s suffixList) PublicSuffix(domain string) string {

	return ""
}

func (s suffixList) String() string {
	return ""
}

type WikiClient struct {
	username string
	password string
	client   *http.Client
	token    string
}

func Wiki(username string, password string) WikiClient {
	suffixList := suffixList{}
	cookieOptions := cookiejar.Options{suffixList}
	cookieJar, _ := cookiejar.New(&cookieOptions)
	webClient := http.Client{Jar: cookieJar, Timeout: time.Second * 10}
	token := getToken(&webClient, "login")

	parameters := fmt.Sprintf("?action=login&lgname=%s&lgpassword=%s&lgtoken=%s&format=json", username, password, token)
	req, err := http.NewRequest("POST", wiki+parameters, nil)
	if err != nil {
		log.Panicln(err)
	}

	_, err = webClient.Do(req)
	if err != nil {
		log.Panicln(err)
	}

	return WikiClient{username, password, &webClient, token}
}

func getToken(client *http.Client, tokenType string) string {
	parameters := fmt.Sprintf(`?action=query&meta=tokens&type=%s&format=json`, tokenType)
	req, err := http.NewRequest("POST", wiki+parameters, nil)
	if err != nil {
		log.Panicln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	str, err := jsonparser.GetString(bytes, "query", "tokens", fmt.Sprintf("%stoken", tokenType))
	if err != nil {
		log.Panicln(err)
	}
	return str
}
