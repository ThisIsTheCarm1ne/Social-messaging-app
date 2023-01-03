package PostHandler

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type post struct {
	Uid      int
	FatherId int
	Title    string
	Content  string
	Date     string
	Pic      string
	Tag      string
}
type lotsOfPosts struct {
	SinglePost []post
	Comment    []post
	Theme      []post
	Popular    []post
}
type MTV struct {
	key   int
	value int
}

func GetPost(postId string, DB *sql.DB, pageId int) lotsOfPosts {

	queryRow := "SELECT post_id, title, content, date, IFNULL(tag, \"\"), IFNULL(pic, \"\") FROM post WHERE post_id=" + postId

	var query string

	if pageId == 0 {
		query = `SELECT comment_id,
				title,
				content,
				date,
				pic FROM comment WHERE postId=` + postId + ` ORDER BY comment_id DESC LIMIT 50`
	} else if pageId < 0 {
		pageId = pageId*-1 + 49
		fmt.Println(pageId)
		query = `SELECT comment_id,
				title,
				content,
				date,
				pic 
				FROM comment WHERE postId=` + postId + ` AND comment_id <= ` + strconv.Itoa(pageId) + ` ORDER BY comment_id DESC LIMIT 50`
	} else {
		query = `SELECT comment_id,
				title,
				content,
				date,
				pic FROM comment WHERE postId=` + postId + ` AND comment_id <= ` + strconv.Itoa(pageId) + ` ORDER BY comment_id DESC LIMIT 50`
	}

	parseThis := lotsOfPosts{Theme: queryFromDB(query, DB), Comment: QueryRowFromDB(queryRow, DB), SinglePost: GetMostPopular(DB)}
	return parseThis
}

func GetThemes(DB *sql.DB, pageId int) lotsOfPosts {

	var query string
	// This query should work - SELECT rowid, title, content, date FROM theme WHERE rowid>=30 ORDER BY rowid DESC LIMIT=?;

	//Creates a query based of pageId variable
	//If pageId is negative, then its call to make query for previous themes
	//If positive, then for next themes
	//If 0, then default. Select newest

	if pageId == 0 {
		query = "SELECT theme_id, title, content, date, pic FROM theme ORDER BY theme_id DESC LIMIT 50"
	} else if pageId < 0 {
		pageId = pageId*-1 + 49
		fmt.Println(pageId)
		query = `SELECT theme_id,
				title,
				content,
				date,
				pic FROM theme WHERE theme_id <= ` + strconv.Itoa(pageId) + ` ORDER BY theme_id DESC LIMIT 50`
	} else {
		query = `SELECT theme_id,
				title,
				content,
				date,
				pic FROM theme WHERE theme_id <= ` + strconv.Itoa(pageId) + ` ORDER BY theme_id DESC LIMIT 50`
	}

	parseThis := lotsOfPosts{Theme: queryFromDB(query, DB), SinglePost: GetMostPopular(DB)}
	return parseThis
}

func GetTheme(themeId string, DB *sql.DB, pageId int) lotsOfPosts {

	queryRow := "SELECT rowid, title, content, date, IFNULL(tag, \"\"), IFNULL(pic, \"\") FROM theme WHERE rowid=" + themeId

	var query string

	if pageId == 0 {
		query = `SELECT post_id,
				title,
				content,
				date,
				pic FROM post WHERE themeId=` + themeId + ` ORDER BY post_id DESC LIMIT 50`
	} else if pageId < 0 {
		pageId = pageId*-1 + 49
		fmt.Println(pageId)
		query = `SELECT post_id,
				title,
				content,
				date,
				pic FROM post WHERE themeId=` + themeId + ` AND post_id <= ` + strconv.Itoa(pageId) + ` ORDER BY post_id DESC LIMIT 50`
	} else {
		query = `SELECT post_id,
				title,
				content,
				date,
				pic FROM post WHERE themeId=` + themeId + ` AND post_id <= ` + strconv.Itoa(pageId) + ` ORDER BY post_id DESC LIMIT 50`
	}

	parseThis := lotsOfPosts{Theme: queryFromDB(query, DB), SinglePost: GetMostPopular(DB), Comment: QueryRowFromDB(queryRow, DB)}
	return parseThis
}

func QueryRowFromDB(query string, DB *sql.DB) []post {

	row := DB.QueryRow(query)
	var (
		Uid     int
		Title   string
		Content string
		Date    string
		Pic     string
		Tag     string
	)

	err := row.Scan(&Uid, &Title, &Content, &Date, &Tag, &Pic)
	fmt.Println(err)
	fmt.Println("Uid- ", Uid, "Title-", Title, "Content- ", Content, "Pic- ", Pic)

	p := make([]post, 0)
	p = append(p, post{Uid: Uid, Title: Title, Content: Content, Date: Date, Tag: Tag, Pic: Pic})

	return p

}
func queryFromDB(query string, DB *sql.DB) []post {

	rows, err := DB.Query(query)

	fmt.Println(err)

	//Collects and then sends data from db

	themes := make([]post, 0)
	for rows.Next() {
		singleTheme := post{}
		err = rows.Scan(&singleTheme.Uid, &singleTheme.Title, &singleTheme.Content, &singleTheme.Date, &singleTheme.Pic)
		themes = append(themes, singleTheme)
	}
	rows.Close()

	return themes
}

func DeleteContent(kind string, id string, DB *sql.DB) {

	var output []post
	var approved []post

	query := `SELECT rowid,
			 title,
			 content,
			 date,
			 agent
			 FROM reports WHERE Id=` + id

	output = queryFromDB(query, DB)

	fmt.Println("output-", output)

	//check if more than 15 report are real and by a different people

	approved = approveReports(output)

	//Delete reported stuff and then clean up
	if len(approved) >= 1 {
		switch kind {
		case "theme":
			queryDel := "DELETE FROM theme WHERE theme_id=?"
			stmt, err := DB.Prepare(queryDel)
			fmt.Println(err)
			res, err := stmt.Exec(id)
			fmt.Println(err)
			affect, err := res.RowsAffected()
			fmt.Println(affect, err)
		case "post":
			queryDel := "DELETE FROM post WHERE post_id=?"
			stmt, err := DB.Prepare(queryDel)
			fmt.Println(err)
			res, err := stmt.Exec(id)
			fmt.Println(err)
			affect, err := res.RowsAffected()
			fmt.Println(affect, err)
		case "comment":
			queryDel := "DELETE FROM comment WHERE comment_id=?"
			stmt, err := DB.Prepare(queryDel)
			fmt.Println(err)
			res, err := stmt.Exec(id)
			fmt.Println(err)
			affect, err := res.RowsAffected()
			fmt.Println(affect, err)
		default:
			fmt.Println("ERROR - type is incorrect or empty!")
		}
	}
	queryDel := "DELETE FROM reports WHERE Id=?"
	stmt, err := DB.Prepare(queryDel)
	fmt.Println(err)
	res, err := stmt.Exec(id)
	fmt.Println(err)
	affect, err := res.RowsAffected()
	fmt.Println(affect, err)

}

func approveReports(output []post) []post {
	var approved []post
	timeFormat := "1/2/2006, 15:04:05 PM"

	for i := 0; i < (len(output) - 1); i++ {
		tMain, err := time.Parse(timeFormat, output[i].Date)
		tPlus := tMain.Add(time.Minute * 1) //Change it to- * 15
		t, err := time.Parse(timeFormat, output[i+1].Date)
		fmt.Println(err)
		if tPlus.Before(t) == true && output[i].Pic != output[i+1].Pic {
			approved = append(approved, output[i])
		}
	}
	fmt.Println("-*Approved*-")

	for i := 0; i < len(approved); i++ {
		fmt.Println(approved[i])
	}

	return approved
}

func GetMostPopular(DB *sql.DB) []post {
	//So... you need somehow pick the most popular theme in 24-hours
	//Welp, i think that you should make a query into db with themes
	//To check for themes that were made in last 24-hours
	//After that made a query into db with post
	//And check in what theme most of the post
	//Yes, right. Somehow with id's of themes
	//And calculate the most common theme id
	//Just use time.Now()

	date := time.Now().Format("1/2/2006")
	var query string
	//Picking themes with the same date

	query = "SELECT theme_id, title, content, date, pic FROM theme WHERE date LIKE " + "'%" + date + "%'"

	themes := queryFromDB(query, DB)
	//Making an array of themes ids

	var ids []string
	for i := 0; i < len(themes); i++ {
		ids = append(ids, strconv.Itoa(themes[i].Uid))
	}

	//Picking post from already picked dates
	queryP := "SELECT post_id, title, content, themeId FROM post WHERE themeId in (" + strings.Join(ids, ", ") + ")"
	//rowsP, err := DB.Query("SELECT rowid, title, content, themeId FROM posts WHERE themeId in (?)", strings.Join(ids, ", "))
	rowsP, err := DB.Query(queryP)

	fmt.Println(err)

	posts := make([]post, 0)
	for rowsP.Next() {
		Post := post{}
		err = rowsP.Scan(&Post.Uid, &Post.Title, &Post.Content, &Post.FatherId)
		posts = append(posts, Post)
	}

	rowsP.Close()

	popularThemes := make([]post, 0)

	if len(posts) > 0 {
		commons := contains(posts)

		for i := 0; i < len(commons); i++ {
			for a := 0; a < len(themes); a++ {
				if commons[i].key == themes[a].Uid {
					popularThemes = append(popularThemes, themes[a])
				}
			}
		}
	}
	return popularThemes

}

func contains(s []post) []MTV {
	commonsD := make([]MTV, 0)
	commons := make([]MTV, 0)
	for i := 0; i < len(s); i++ {
		b := 0
		for a := 0; a < len(s); a++ {
			if s[i].FatherId == s[a].FatherId {
				b = b + 1
			}
		}
		if b > 0 {
			Common := MTV{s[i].FatherId, b}
			commonsD = append(commonsD, Common)
		}
	}

	keys := make(map[MTV]bool)

	for _, entry := range commonsD {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			commons = append(commons, entry)
		}
	}

	sort.Slice(commons, func(d, j int) bool {
		if commons[d].value != commons[j].value {
			return commons[d].value > commons[j].value
		}
		// 2. only when value is the same - sort by key
		return commons[d].key < commons[j].key
	})

	return commons
}
