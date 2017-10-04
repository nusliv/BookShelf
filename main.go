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
	SearchResult struct {
		Title  string `xml:"title,attr"`
		Author string `xml:"author, attr"`
		Year   string `xml:"hyr,attr"`
		ID     string `xml:"owi,attr"`
	}
	ClassifySearchResponse struct {
		Results []SearchResult `xml:"works>work"`
	}
)

// search will search for a query and return a list of results
func search(query string) ([]SearchResult, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get("http://classify.oclc.org/classify2/Classify?&summary=true&title=" + url.QueryEscape(query)); err != nil {
		return []SearchResult{}, err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return []SearchResult{}, err
	}

	var c ClassifySearchResponse
	err = xml.Unmarshal(body, &c)

	return c.Results, err
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

	// Search request
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

	log.Println(http.ListenAndServe(":8080", nil))
}
