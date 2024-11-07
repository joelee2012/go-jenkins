package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

type ApiJsonOpts struct {
	Tree  string
	Depth int
}

func (o *ApiJsonOpts) Encode() string {
	v := url.Values{}
	if o.Tree != "" {
		v.Add("tree", o.Tree)
	}
	v.Add("depth", strconv.Itoa(o.Depth))
	return v.Encode()
}

// Bind jenkins JSON data to any type,
//
//	// bind json data to map
//	data := make(map[string]string)
//	jenkins.ApiJson(&data, &ApiJsonOpts{"tree":"description"})
//	fmt.Println(data["description"])
func (i *Item) ApiJson(v any, opts *ApiJsonOpts) error {
	return unmarshalApiJson(i, v, opts)
}

func unmarshalApiJson(r Requester, v any, opts *ApiJsonOpts) error {
	entry := "api/json"
	if opts != nil {
		entry = "api/json?" + opts.Encode()
	}
	resp, err := r.Request("GET", entry, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
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
