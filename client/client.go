package client

import (
	"database/sql"
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
	_ "github.com/lib/pq"
)

const wiki string = "https://wiki.teamfortress.com/w/api.php"

var ignoreTemplates []string
var ignoreParameters []string

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
	database *sql.DB
}

type WikiPage struct {
	namespace int64
	article   string
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

	databaseData, err := ioutil.ReadFile("db.txt")
	if err != nil {
		log.Fatalf("Error opening db.txt\n%s", err)
	}

	database, databaseErr := sql.Open("postgres", string(databaseData))
	if databaseErr != nil {
		log.Fatalf("Error opening database.\n%s", databaseErr)
	}

	err = database.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database. This is most likely a problem with db.txt.\n%s", err)
	}

	statement, tableErr := database.Prepare("CREATE TABLE IF NOT EXISTS wikipages (title VARCHAR(255) PRIMARY KEY, points FLOAT, lastseen DATE)")
	statement.Exec()
	if tableErr != nil {
		log.Fatalf("Error creating table.\n%s", tableErr)
	}

	client := WikiClient{username, password, &webClient, token, make(chan []byte), database}

	defer client.RequestLoop()

	templatesFile, err := ReadCsv("ignored_templates.csv")
	if err != nil {
		log.Printf("! Error reading ignored_templates.csv\n%s", err)
		ignoreTemplates = []string{}
	} else {
		ignoreTemplates = templatesFile
	}

	parametersFile, err := ReadCsv("ignored_parameters.csv")
	if err != nil {
		log.Printf("! Error reading ignored_parameters.csv\n%s", err)
		ignoreParameters = []string{}
	} else {
		ignoreParameters = parametersFile
	}

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
		defer w.database.Close()
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

func (w *WikiClient) GetArticles(titles []string) (map[string]WikiPage, error) {
	parameters := fmt.Sprintf(`?action=query&prop=revisions&titles=%s&rvprop=content&format=json`, url.PathEscape(strings.Join(titles, "|")))
	api := w.DoRequest(parameters)

	if len(api) == 0 {
		return nil, errors.New("[GetArticles] Error making API request")
	}
	content := make(map[string]WikiPage)

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
		namespace, err := jsonparser.GetInt(value, "ns")
		rev, _, _, err := jsonparser.Get(value, "revisions")
		if err != nil {
			// pagina não existe (provavelmente)
			return nil
		}
		_, err = jsonparser.ArrayEach(rev, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
			page, _ := jsonparser.GetString(value, "*")
			content[title] = WikiPage{namespace: namespace, article: page}
		})
		return err
	}
	err = jsonparser.ObjectEach(pages, getPage)
	if err != nil {
		return nil, fmt.Errorf("[GetArticles] Error iterating over pages->\n\t%s", err)
	}

	return content, nil
}

func (w WikiClient) RenderPage() string {
	pages, err := w.getDBEntries(false)
	if err != nil {
		log.Fatalf("[RenderPages] Error getting DB entries:\n%s", err)
	}

	return "fuck"
}
