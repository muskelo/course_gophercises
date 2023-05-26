package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

/* 
Shortcuts ------------------------------------------------------------------
*/

type Shortcut struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// type to create Shortner Handler
// fill in with Parse methods and create Handler
type ShortcutsList []Shortcut

func (list ShortcutsList) GetUrl(Path string) string {
	for i := range list {
		if list[i].Path == Path {
			return list[i].URL
		}
	}
	return ""
}

func (list ShortcutsList) Handler(fallback http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Url := list.GetUrl(r.URL.Path)
		if Url == "" {
			fallback(w, r)
			return
		}
		http.Redirect(w, r, Url, http.StatusTemporaryRedirect)
	}
}

// parse map where path is key, url is value
func (list ShortcutsList) ParseMap(Map map[string]string) ShortcutsList {
	for k, v := range Map {
		list = append(list, Shortcut{k, v})
	}
	return list
}

// parse yaml file in format
//
//   - path: /some-path
//     url : http://my-url.com
func (list ShortcutsList) ParseYaml(fn string) (ShortcutsList, error) {
	// open
	file, err := os.Open(fn)
	if err != nil {
		return list, err
	}
	// parse
	readed := make(ShortcutsList, 0)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&readed)
	if err != nil {
		return list, err
	}
	// return
	return append(list, readed...), nil
}

// parse json file in format
//
//	[{
//		  "path": "/some-path",
//		  "url" : "http://my-url.com"
//	}]
func (list ShortcutsList) ParseJSON(fn string) (ShortcutsList, error) {
	// open
	file, err := os.Open(fn)
	if err != nil {
		return list, err
	}
	// parse
	readed := make(ShortcutsList, 0)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&readed)
	if err != nil {
		return list, err
	}
	return append(list, readed...), nil
}

func DefaultFallback(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

/* 
Main ------------------------------------------------------------------
*/

// flag.Value to store path:url values
type MapValue map[string]string

func (mapValue MapValue) String() string {
	return ""
}
func (mapValue MapValue) Set(value string) error {
	splited := strings.SplitN(value, ":", 2)
	if len(splited) != 2 {
		return fmt.Errorf("Expected value in format 'key:value'")
	}
	mapValue[splited[0]] = splited[1]
	return nil
}

type Flags struct {
	YAML string
	JSON string
	Map  MapValue
}

func parseFlags() Flags {
	flags := Flags{
		Map: make(MapValue),
	}
	flag.StringVar(&flags.YAML, "yaml", "", "yaml file with shortcuts")
	flag.StringVar(&flags.JSON, "json", "", "json file with shortcuts")
	flag.Var(flags.Map, "map", "path:url map")
	flag.Parse()
	return flags
}

func main() {
	flags := parseFlags()

	// build
	shortcuts := make(ShortcutsList, 0)
	if len(flags.Map) > 0 {
		shortcuts = shortcuts.ParseMap(flags.Map)
	}
	if flags.YAML != "" {
		var err error
		shortcuts, err = shortcuts.ParseYaml(flags.YAML)
		if err != nil {
			panic(err)
		}
	}
	if flags.JSON != "" {
		var err error
		shortcuts, err = shortcuts.ParseJSON(flags.JSON)
		if err != nil {
			panic(err)
		}
	}

	// run
	err := http.ListenAndServe("0.0.0.0:5000", shortcuts.Handler(DefaultFallback))
	if err != nil {
		panic(err)
	}
}
