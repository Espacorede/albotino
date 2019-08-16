package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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
	channel  chan []byte
}

func Wiki(username string, password string) *WikiClient {
	suffixList := suffixList{}
	cookieOptions := cookiejar.Options{PublicSuffixList: suffixList}
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

	client := WikiClient{username, password, &webClient, token, make(chan []byte)}

	defer client.RequestLoop()

	return &client
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

func (w *WikiClient) RequestLoop() {
	go func() {
		for {
			request := <-w.channel
			resp, err := w.WikiAPIRequest(request)
			if err != nil {
				log.Printf("[RequestLoop] Error on API request: %s", err)
				w.channel <- make([]byte, 0)
			} else {
				w.channel <- resp
			}
			time.Sleep(time.Second)
		}
	}()
}

func (w *WikiClient) WikiAPIRequest(parameters []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", wiki+string(parameters), nil)
	if err != nil {
		return nil, err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (w *WikiClient) DoRequest(parameters string) []byte {
	w.channel <- []byte(parameters)
	return <-w.channel
}

func (w *WikiClient) GetArticles(titles []string) (map[string][]byte, error) {
	parameters := fmt.Sprintf(`?action=query&prop=revisions&titles=%s&rvprop=content&format=json`, url.PathEscape(strings.Join(titles, "|")))
	api := w.DoRequest(parameters)
	if len(api) == 0 {
		return nil, errors.New("[GetArticles] Error making API request")
	}
	content := make(map[string][]byte)

	pages, _, _, err := jsonparser.Get(api, "query", "pages")
	if err != nil {
		log.Println(titles)
		log.Println(string(api))
		return nil, fmt.Errorf("[GetArticles] Error getting json pages->\n\t%s", err)
	}

	getPage := func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		title, err := jsonparser.GetString(value, "title")
		if err != nil {
			return fmt.Errorf("[GetArticles] Error getting json page title->\n\t%s", err)
		}
		rev, _, _, err := jsonparser.Get(value, "revisions")
		if err != nil {
			// pagina não existe (provavelmente)
			return nil
		}
		_, err = jsonparser.ArrayEach(rev, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
			page, _, _, _ := jsonparser.Get(value, "*")
			content[title] = page
		})
		return err
	}
	err = jsonparser.ObjectEach(pages, getPage)
	if err != nil {
		return nil, fmt.Errorf("[GetArticles] Error iterating over pages->\n\t%s", err)
	}

	return content, nil
}
