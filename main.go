package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	PostHandler "messager/src"

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

var DB *sql.DB

func postHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseGlob("views/postPage.html")

	var fullName string

	r.ParseMultipartForm(10 << 20)

	file, handler, errPic := r.FormFile("commentPic")

	pageIdStr := r.FormValue("pageId")
	pageId, err := strconv.Atoi(pageIdStr)
	fmt.Println(err)

	//SQLinjection protection
	postIdStr := r.FormValue("postId")
	postId, err := strconv.Atoi(postIdStr)
	postIdClean := strconv.Itoa(postId)

	if len(postIdStr) != len(postIdClean) {
		fmt.Println("This is triggered")
		t, _ = template.ParseGlob("views/mainPage.html")
		t.Execute(w, PostHandler.GetThemes(DB, 0))
		return
	}
	dateSlice := strings.Split(r.FormValue("commentDate"), ",")

	fmt.Println(dateSlice[0])

	// IF method POST & both field posts data, write data to the database
	if len(r.FormValue("commentName")) > 0 && len(r.FormValue("commentDesc")) > 0 {

		if errPic == nil && strings.Contains(handler.Header["Content-Type"][0], "image") {
			defer file.Close()

			dFile, err := ioutil.TempFile("public/user_data", "file-*.png")

			defer dFile.Close()

			fileBytes, err := ioutil.ReadAll(file)
			fmt.Println(err)

			dFile.Write(fileBytes)
			fName := strings.Split(dFile.Name(), "/")
			fullName = fName[len(fName)-2] + "/" + fName[len(fName)-1]
		}

		stmt, err := DB.Prepare("INSERT INTO comment(title, content, date, pic, postId) values(?, ?, ?, ?, ?)")
		fmt.Println(err)
		res, err := stmt.Exec(r.FormValue("commentName"), r.FormValue("commentDesc"), r.FormValue("commentDate"), fullName,
			postIdStr)
		fmt.Println(res, err)
		//Something like Post/Redirect/Get
		http.Redirect(w, r, "/theme/post?postId="+postIdStr, http.StatusFound)
	}

	t.Execute(w, PostHandler.GetPost(postIdStr, DB, pageId))
}

func themeHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseGlob("views/themePage.html")

	var fullName string

	r.ParseMultipartForm(10 << 20)
	file, handler, errPic := r.FormFile("postPic")

	//themeId is needed to display right posts
	themeIdStr := r.FormValue("themeId")
	themeId, err := strconv.Atoi(themeIdStr)
	themeIdClean := strconv.Itoa(themeId)

	if len(themeIdStr) != len(themeIdClean) {
		fmt.Println("This is triggered")
		t, _ = template.ParseGlob("views/mainPage.html")
		t.Execute(w, PostHandler.GetThemes(DB, 0))
		return
	}
	//pageId is needed to "swipe" through posts
	pageIdStr := r.FormValue("pageId")
	pageId, err := strconv.Atoi(pageIdStr)

	fmt.Println(err)

	dateSlice := strings.Split(r.FormValue("themeDate"), ",")

	fmt.Println(dateSlice[0])

	// IF method POST & both field posts data, write data to the database
	if len(r.FormValue("postName")) > 0 && len(r.FormValue("postDesc")) > 0 {

		if errPic == nil && strings.Contains(handler.Header["Content-Type"][0], "image") {
			defer file.Close()

			dFile, err := ioutil.TempFile("public/user_data", "file-*.png")

			defer dFile.Close()

			fileBytes, err := ioutil.ReadAll(file)
			fmt.Println(err)

			dFile.Write(fileBytes)
			fName := strings.Split(dFile.Name(), "/")
			fullName = fName[len(fName)-2] + "/" + fName[len(fName)-1]
		}

		stmt, err := DB.Prepare("INSERT INTO post(title, content, date, pic, themeId) values(?, ?, ?, ?, ?)")
		fmt.Println(err)
		res, err := stmt.Exec(r.FormValue("postName"), r.FormValue("postDesc"), r.FormValue("postDate"), fullName,
			themeId)
		fmt.Println(res, err)
		//Something like Post/Redirect/Get
		http.Redirect(w, r, "/theme?themeId="+themeIdStr, http.StatusFound)
	}

	t.Execute(w, PostHandler.GetTheme(themeIdStr, DB, pageId))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseGlob("views/searchPage.html")

	r.ParseForm()

	fmt.Println("searchOptions-", r.FormValue("searchOptions"))

	search := r.FormValue("search")

	query := "SELECT theme_id, title, content, date, pic FROM theme WHERE title LIKE " + "'%" + search + "%'" + " ORDER BY theme_id DESC"
	fmt.Println(query)
	rows, err := DB.Query(query)
	fmt.Println(err)

	themes := make([]post, 0)

	for rows.Next() {
		singleTheme := post{}
		err = rows.Scan(&singleTheme.Uid, &singleTheme.Title, &singleTheme.Content, &singleTheme.Date, &singleTheme.Pic)
		themes = append(themes, singleTheme)
	}
	rows.Close()
	parseThis := lotsOfPosts{Theme: themes}

	fmt.Println(parseThis)

	t.Execute(w, parseThis)
}
func reportHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("views/reportPage.html")

	r.ParseForm()

	id := r.FormValue("id")
	kind := r.FormValue("type")
	reportTitle := r.FormValue("reportName")
	reportDesc := r.FormValue("reportDesc")
	reportAgent := r.FormValue("reportAgent")
	reportDate := r.FormValue("reportDate")

	query := "SELECT " + kind + "_id, title, content, date, IFNULL(tag, \"\"), IFNULL(pic, \"\") FROM " + kind + " WHERE " + kind + "_id=" + id

	p := PostHandler.QueryRowFromDB(query, DB)

	pLocal := make([]post, 0)

	pLocal = append(pLocal, post{Uid: p[0].Uid, Title: p[0].Title, Content: p[0].Content, Date: p[0].Date, Tag: p[0].Tag, Pic: p[0].Pic})

	parseThis := lotsOfPosts{Comment: pLocal}

	if len(reportTitle) > 0 && len(reportDesc) > 0 {
		stmt, err := DB.Prepare("INSERT INTO reports(title, content, date, agent, Id) values(?, ?, ?, ?, ?)")
		fmt.Println(err)
		res, err := stmt.Exec(reportTitle, reportDesc, reportDate, reportAgent, id)
		fmt.Println(res, err)
		PostHandler.DeleteContent(kind, id, DB)
		http.Redirect(w, r, "/report?type="+kind+"&"+"id="+id, http.StatusFound)
	}

	t.Execute(w, parseThis)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseGlob("views/mainPage.html")

	var fullName string

	r.ParseMultipartForm(10 << 20)

	file, handler, errPic := r.FormFile("themePic")

	pageIdStr := r.FormValue("pageId")

	pageId, err := strconv.Atoi(pageIdStr)
	fmt.Println(err)

	dateSlice := strings.Split(r.FormValue("themeDate"), ",")

	fmt.Println(dateSlice[0])

	// IF method POST & both field posts data, write data to the database
	if len(r.FormValue("themeName")) > 0 && len(r.FormValue("themeDesc")) > 0 {

		if errPic == nil && strings.Contains(handler.Header["Content-Type"][0], "image") {
			defer file.Close()

			dFile, err := ioutil.TempFile("public/user_data", "file-*.png")

			defer dFile.Close()

			fileBytes, err := ioutil.ReadAll(file)
			fmt.Println(err)

			dFile.Write(fileBytes)
			fName := strings.Split(dFile.Name(), "/")
			fullName = fName[len(fName)-2] + "/" + fName[len(fName)-1]
		}

		stmt, err := DB.Prepare("INSERT INTO theme(title, content, date, pic) values(?, ?, ?, ?)")
		fmt.Println(err)
		res, err := stmt.Exec(r.FormValue("themeName"), r.FormValue("themeDesc"), r.FormValue("themeDate"), fullName)
		fmt.Println(res, err)
		//Something like Post/Redirect/Get
		http.Redirect(w, r, "/", http.StatusFound)
	}

	t.Execute(w, PostHandler.GetThemes(DB, pageId))
}

func main() {
	db, _ := sql.Open("sqlite3", "file:themeDatabase.db?_foreign_keys=on")
	DB = db

	fs := http.FileServer(http.Dir("public/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/theme/post", postHandler)
	http.HandleFunc("/theme", themeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/report", reportHandler)

	fmt.Println("Server has been successfully launched")
	fmt.Println("And hosted on port 5000!")

	http.ListenAndServe(":5000", nil)
}
