package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

var categories = []string{
	"scout cosmetics",
	"soldier cosmetics",
	"pyro cosmetics",
	"demoman cosmetics",
	"heavy cosmetics",
	"engineer cosmetics",
	"medic cosmetics",
	"sniper cosmetics",
	"spy cosmetics",
	"multiclass cosmetics",
	"allclass cosmetics",
	"weapons",
	"other"}

type PageEntry struct {
	Title      string
	points     []sql.NullInt64
	lastupdate time.Time
	category   string
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

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS wikipages (title VARCHAR(255) PRIMARY KEY, points INT[], lastseen DATE, category varchar(255))")

	upsert, err = database.Prepare(`
		INSERT INTO wikipages (title, points, lastseen, category) 
		VALUES ($1, $2, current_date, $3)
		ON CONFLICT (title) DO UPDATE 
		SET points = excluded.points, 
			lastseen = excluded.lastseen,
			category = excluded.category;`)
	if err != nil {
		return nil, err
	}

	allEntries, err = database.Prepare(`
		SELECT title, points, lastseen, category
		FROM wikipages
		ORDER BY title ASC;`)
	if err != nil {
		return nil, err
	}

	outOfDate, err = database.Prepare(`
	SELECT title, points, lastseen, category
	FROM wikipages
	WHERE lastseen < CURRENT_DATE - interval '1 week'
	ORDER BY title ASC;`)
	if err != nil {
		return nil, err
	}

	return database, err
}

func upsertDBEntry(title string, points []int64, category string) error {
	sqlArray := make([]sql.NullInt64, len(points))

	for index, num := range points {
		if num == -1 {
			sqlArray[index] = sql.NullInt64{Int64: 0, Valid: false}
		} else {
			sqlArray[index] = sql.NullInt64{Int64: num, Valid: true}
		}
	}

	_, err := upsert.Exec(title, pq.Array(sqlArray), category)

	return err
}

func GetDBEntries(outdated bool) (map[string][]PageEntry, error) {
	entries := map[string][]PageEntry{}

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
		var category string
		err = rows.Scan(&title, pq.Array(&points), &lastseen, &category)
		if err != nil {
			return nil, err
		}
		pageObj := PageEntry{title, points, lastseen, category}
		entries[category] = append(entries[category], pageObj)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func RenderPages() {
	page, err := ioutil.ReadFile("page.txt")
	if err != nil {
		log.Printf("[RenderPage] Error reading page.txt->\n\t%s\n", err)
		return
	}

	tables, err := RenderTables()
	if err != nil {
		log.Printf("[RenderPage] Error rendering tables->\n\t%s\n", err)
	} else {
		for key, value := range tables {
			err = ioutil.WriteFile(fmt.Sprintf("temp/%s.txt", key), []byte(fmt.Sprintf(string(page), value)), 0644)
			if err != nil {
				log.Printf("[RenderPage] Error saving %s->\n\t%s\n", key, err)
			}
		}
	}
}

func RenderTables() (map[string]string, error) {
	tables := map[string]string{}

	categories, err := GetDBEntries(false)
	if err != nil {
		return tables, err
	}

	for category, pages := range categories {
		var sb strings.Builder
		for _, page := range pages {
			var pb strings.Builder
			pb.WriteString(fmt.Sprintf("|-\n\t| [[%s]] ", page.Title))

			for index, language := range page.points {
				var pointsString string
				if language.Valid {
					pointsString = fmt.Sprintf("%d", language.Int64)
				} else {
					pointsString = "N/A"
				}
				pb.WriteString(fmt.Sprintf("|| [[%s|%s]] ", page.Title+"/"+languages[index], pointsString))
			}

			pb.WriteString(fmt.Sprintf("\n\t| data-sort-value=\"%s\" | %s \n", page.lastupdate, page.lastupdate.Format(time.RFC3339)[:10]))
			sb.WriteString(pb.String())
		}
		tables[category] = sb.String()
	}
	return tables, nil
}
