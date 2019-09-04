package client

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

type PageEntry struct {
	Title      string
	points     []sql.NullInt64
	lastupdate time.Time
}

var database *sql.DB
var upsert *sql.Stmt
var allEntries *sql.Stmt
var outOfDate *sql.Stmt

func SetupDatabase(host string, port string, user string, password string, db string) (*sql.DB, error) {
	dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, db)

	var err error

	database, err = sql.Open("postgres", dbString)
	if err != nil {
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS wikipages (title VARCHAR(255) PRIMARY KEY, points INT[], lastseen DATE)")

	upsert, err = database.Prepare(`
		INSERT INTO wikipages (title, points, lastseen) 
		VALUES ($1, $2, current_date)
		ON CONFLICT (title) DO UPDATE 
		SET points = excluded.points, 
			lastseen = excluded.lastseen;`)
	if err != nil {
		return nil, err
	}

	allEntries, err = database.Prepare(`
		SELECT title, points, lastseen
		FROM wikipages`)
	if err != nil {
		return nil, err
	}

	outOfDate, err = database.Prepare(`
	SELECT title, points, lastseen
	FROM wikipages
	WHERE lastseen < CURRENT_DATE - interval '1 week'`)
	if err != nil {
		return nil, err
	}

	return database, err
}

func upsertDBEntry(title string, points []int64) error {
	sqlArray := make([]sql.NullInt64, len(points))

	for index, num := range points {
		if num == -1 {
			sqlArray[index] = sql.NullInt64{Int64: 0, Valid: false}
		} else {
			sqlArray[index] = sql.NullInt64{Int64: num, Valid: true}
		}
	}

	_, err := upsert.Exec(title, pq.Array(sqlArray))

	return err
}

func GetDBEntries(outdated bool) ([]PageEntry, error) {
	entries := []PageEntry{}

	var rows *sql.Rows
	var err error

	if outdated {
		rows, err = outOfDate.Query()
	} else {
		rows, err = allEntries.Query()
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		var points []sql.NullInt64
		var lastseen time.Time
		err = rows.Scan(&title, &points, &lastseen)
		if err != nil {
			return nil, err
		}

		entries = append(entries, PageEntry{title, points, lastseen})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func RenderPage() string {
	page, err := ioutil.ReadFile("page.txt")
	if err != nil {
		log.Printf("[RenderPage] Error reading page.txt:\n%s", err)
		return ""
	}
	return fmt.Sprintf(string(page), RenderTable())
}

func RenderTable() string {
	pages, err := GetDBEntries(false)
	if err != nil {
		log.Printf("[RenderTable] Error getting DB entries:\n%s", err)
		return ""
	}

	var sb strings.Builder

	for _, page := range pages {
		var pb strings.Builder
		pb.WriteString(fmt.Sprintf("|- | [[%s]] ", page.Title))

		for index, language := range page.points {
			var pointsString string
			if language.Valid {
				pointsString = string(language.Int64)
			} else {
				pointsString = "N/A"
			}
			pb.WriteString(fmt.Sprintf("|| [[%s|%s]]", page.Title+"/"+languages[index], pointsString))
		}

		pb.WriteString(fmt.Sprintf("|| %s ", page.lastupdate))
		sb.WriteString(pb.String())
	}
	return sb.String()
}
