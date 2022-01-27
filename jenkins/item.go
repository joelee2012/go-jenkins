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
	URL    string
	Class  string
	client *Client
}

type ReqParams = req.Param

func NewItem(url, class string, client *Client) *Item {
	url = appendSlash(url)
	return &Item{URL: url, Class: parseClass(class), client: client}
}

func (i *Item) BindAPIJson(params ReqParams, v interface{}) error {
	resp, err := i.Request("GET", "api/json", params)
	if err != nil {
		return err
	}
	return resp.ToJSON(v)
}

func (i *Item) Request(method, entry string, vs ...interface{}) (*req.Resp, error) {
	return i.client.Do(method, i.URL+entry, vs...)
}

func (i *Item) String() string {
	return fmt.Sprintf("<%s: %s>", i.Class, i.URL)
}

func (i *Item) WithContext(ctx context.Context) *Item {
	i.client.WithContext(ctx)
	return i
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
