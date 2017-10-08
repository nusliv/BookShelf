package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

type (
	page struct {
		Name     string
		DBStatus bool
	}
	// SearchResult stores the xml parsed result for a search result
	SearchResult struct {
		Title  string `xml:"title,attr"`
		Author string `xml:"author, attr"`
		Year   string `xml:"hyr,attr"`
		ID     string `xml:"owi,attr"`
	}

	// ClassifySearchResponse holds the search results
	ClassifySearchResponse struct {
		Results []SearchResult `xml:"works>work"`
	}

	// ClassifyBookResponse holds the nested struct to hold classifications
	ClassifyBookResponse struct {
		BookData struct {
			Title  string `xml:"title,attr"`
			Author string `xml:"author,attr"`
			ID     string `xml:"owi,attr"`
		} `xml:"work"`
		Classification struct {
			MostPopular string `xml:"sfa,attr"`
		} `xml:"recommendations>ddc>mostPopular"`
	}
)

// find searches for a matching ID
func find(id string) (ClassifyBookResponse, error) {
	var c ClassifyBookResponse
	var body []byte
	var err error

	body, err = classifyAPI("http://classify.oclc.org/classify2/Classify?&summary=true&owi=" + url.QueryEscape(id))
	if err != nil {
		return ClassifyBookResponse{}, err
	}

	if err = xml.Unmarshal(body, &c); err != nil {
		return ClassifyBookResponse{}, err
	}
	return c, nil
}

// search will search for a query and return a list of results
func search(query string) ([]SearchResult, error) {
	var c ClassifySearchResponse
	var body []byte
	var err error

	if body, err = classifyAPI("http://classify.oclc.org/classify2/Classify?&summary=true&title=" + url.QueryEscape(query)); err != nil {
		return []SearchResult{}, err
	}

	err = xml.Unmarshal(body, &c)

	return c.Results, err
}

// classifyAPI helper function for classification
func classifyAPI(url string) ([]byte, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get(url); err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func main() {
	fmt.Println("Application started")
	// Open templates
	// templates := template.Must(template.ParseFiles("templates/index.html"))

	// Open DB
	db, _ := sql.Open("sqlite3", "db/dev.db")
	defer db.Close()

	// Handle index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates := template.Must(template.ParseFiles("src/webSrv/templates/index.html"))
		fmt.Println("Handling request")
		p := &page{Name: "Gopher"}
		if name := r.FormValue("name"); name != "" {
			p.Name = name
		}
		p.DBStatus = db.Ping() == nil
		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle Search request
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		results, err := search(r.FormValue("search"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle Add Books
	http.HandleFunc("/books/add", func(w http.ResponseWriter, r *http.Request) {
		var book ClassifyBookResponse
		var err error

		// Get book ID
		if book, err = find(r.FormValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Check if DB is alive
		if err = db.Ping(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Add book to DB
		if _, err = db.Exec("insert into books (pk, title, author, id, classification) values (?, ?, ?, ?, ?",
			nil, book.BookData.Title, book.BookData.Author, book.BookData.ID, book.Classification.MostPopular); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println(http.ListenAndServe(":8080", nil))
}
