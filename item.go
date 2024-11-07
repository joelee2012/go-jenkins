package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type Item struct {
	URL     string
	Class   string
	jenkins *Jenkins
}

func NewItem(url, class string, jenkins *Jenkins) *Item {
	url = appendSlash(url)
	return &Item{URL: url, Class: parseClass(class), jenkins: jenkins}
}

func (i *Item) ApiJson(v any, opts *ApiJsonOpts) error {
	return unmarshalApiJson(i, v, opts)
}

func (i *Item) Request(method, entry string, body io.Reader) (*http.Response, error) {
	return i.jenkins.doRequest(method, i.URL+entry, body)
}

func (i *Item) String() string {
	return fmt.Sprintf("<%s: %s>", i.Class, i.URL)
}

var delimeter = regexp.MustCompile(`\w+$`)

func parseClass(text string) string {
	return delimeter.FindString(text)
}

func appendSlash(url string) string {
	if strings.HasSuffix(url, "/") {
		return url
	}
	return url + "/"
}

func parseId(url string) int {
	_, base := path.Split(strings.Trim(url, "/"))
	id, _ := strconv.Atoi(base)
	return id
}

func prettyPrintJson(v any) {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))
}
