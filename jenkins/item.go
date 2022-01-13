package jenkins

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req"
)

type Item struct {
	URL     string
	Class   string
	jenkins *Jenkins
}

type ReqParams = req.Param

func NewItem(url, class string, jenkins *Jenkins) *Item {
	url = appendSlash(url)
	return &Item{URL: url, Class: parseClass(class), jenkins: jenkins}
}

func (i *Item) BindAPIJson(params ReqParams, v interface{}) error {
	return doBindAPIJson(i, params, v)
}

func (i *Item) Request(method, entry string, vs ...interface{}) (*req.Resp, error) {
	return doRequest(i.jenkins, method, i.URL+entry, vs...)
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

func prettyPrintJson(v interface{}) {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))
}
