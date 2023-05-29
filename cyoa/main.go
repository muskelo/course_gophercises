package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PageOption struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

// Struct to parse data from json
type Page struct {
	Title   string       `json:"title"`
	Story   []string     `json:"story"`
	Options []PageOption `json:"options"`
}

type Handler struct {
	GeneratedPages []string
	WWWRoot        string
}

// Check that page was generated
func (handler *Handler) PageExist(name string) bool {
	for i := range handler.GeneratedPages {
		if handler.GeneratedPages[i] == name {
			return true
		}
	}
	return false
}

// Parse page data from the json file
func (handler *Handler) parseJSON(path string) (map[string]Page, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	pages := make(map[string]Page)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pages)
	return pages, err
}

// Generate static html pages
func (handler *Handler) GeneratePages(dataPath string, templatePath string) error {
	data, err := handler.parseJSON(dataPath)
	if err != nil {
		return err
	}

	temp, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	for name, page := range data {
		if handler.PageExist(name) {
			return fmt.Errorf("Page with name %v already exist!", name)
		}

		file, err := os.Create(handler.HTMLName(name))
		if err != nil {
			return err
		}
		err = temp.Execute(file, page)

		handler.GeneratedPages = append(handler.GeneratedPages, name)
	}
	return nil
}

// Converts the page name to the path to html file
func (handler *Handler) HTMLName(name string) string {
	return filepath.Join(handler.WWWRoot, name+".html")
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pageName := strings.Trim(r.URL.Path, "/")
	if pageName == "" {
		pageName = "intro"
	}
	if !handler.PageExist(pageName) {
		log.Printf("404, Can't find page with name: %v", pageName)
		w.WriteHeader(404)
		return
	}

	file, err := os.Open(handler.HTMLName(pageName))
	if err != nil {
		log.Printf("500, Error while openening page: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("500, Error while sending page: %v", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("200, Get page %v", pageName)
}

func main() {
	handler := &Handler{
		make([]string, 0),
		"html",
	}
	err := handler.GeneratePages("gopher.json", "page.html")
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe("0.0.0.0:4500", handler)
	if err != nil {
		panic(err)
	}
}
