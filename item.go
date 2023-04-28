package jenkins

import (
	"context"
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
	BaseURL string
	Class   string
	*Client
}


func NewItem(url, class string, client *Client) *Item {
	url = appendSlash(url)
	return &Item{BaseURL: url, Class: parseClass(class), Client: client}
}

func (i *Item) BindAPIJson(params map[string]string, v interface{}) error {
	_, err := i.R().SetQueryParams(params).SetResult(v).Get("api/json")
	return err
}


func (i *Item) String() string {
	return fmt.Sprintf("<%s: %s>", i.Class, i.BaseURL)
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
