package client

type PageEntry struct {
	title  string
	points float64
}

func (w WikiClient) insertDBEntry(title string, points float64) error {
	statement := `
	INSERT INTO wikipages (title, points, lastseen)
	VALUES ($1, $2, current_date)`

	_, err := w.database.Exec(statement, title, points)
	return err
}

func (w WikiClient) updateDBEntry(title string, points float64) error {
	statement := `
	UPDATE wikipages
	SET points = $2, lastseen = CURRENT_DATE
	WHERE id = $1;`

	_, err := w.database.Exec(statement, title, points)
	return err
}

func (w WikiClient) getDBEntries() ([]PageEntry, error) {
	entries := []PageEntry{}

	rows, err := w.database.Query(`
	SELECT title, points
	FROM wikipages
	WHERE lastseen < CURRENT_DATE - interval '1 week'`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		var points float64
		err = rows.Scan(&title, &points)
		if err != nil {
			return nil, err
		}

		entries = append(entries, PageEntry{title, points})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return entries, nil
}
