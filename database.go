package main

func insertIntoDB(title string, points float64, broken []string, wrong []string) error {
	brokenLinks := strings.join(broken, "\n")
	wrongLinks := strings.join(wrong, "\n")

	command, err = database.Prepare("INSERT INTO wikipages (title, points, brokenlinks, wronglinks, lastseen) VALUES (?, ?, ?, ?, ?)")
	command.Exec(title, points, brokenLinks, wrongLinks, time.now())
	
	return err
}
