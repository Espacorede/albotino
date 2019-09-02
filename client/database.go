package client

import (
	"database/sql"
	"fmt"
)

type PageEntry struct {
	title      string
	points     []int
	lastupdate string
}

var Database *sql.DB

func SetupDatabase(host string, port string, user string, password string, db string) error {
	dbString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, db)

	Database, err := sql.Open("postgres", dbString)
	if err != nil {
		return err
	}

	err = Database.Ping()
	if err != nil {
		return err
	}

	statement, err := Database.Prepare("CREATE TABLE IF NOT EXISTS wikipages (title VARCHAR(255) PRIMARY KEY, points INT[], lastseen DATE)")
	statement.Exec()
	return err
}

func upsertDBEntry(title string, points []int) error {
	statement := `
	INSERT INTO the_table (title, points, lastseen) 
	VALUES ($1, $2, current_date)
	ON CONFLICT (title) DO UPDATE 
  	SET points = excluded.points, 
      lastseen = excluded.lastseen;`

	_, err := Database.Exec(statement, title, points)
	return err
}

func getDBEntries(outdated bool) ([]PageEntry, error) {
	entries := []PageEntry{}

	var statement string

	if outdated {
		statement = `
		SELECT title, points, lastseen
		FROM wikipages
		WHERE lastseen < CURRENT_DATE - interval '1 week'`
	} else {
		statement = `
		SELECT title, points, lastseen
		FROM wikipages`
	}

	rows, err := Database.Query(statement)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		var points []int
		var lastseen string
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
