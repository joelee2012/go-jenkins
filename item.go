package jenkins

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

type Item struct {
	Url, Class string
	jenkins    *Jenkins
}

func (i *Item) Request(method, entry string, vs ...interface{}) (*req.Resp, error) {
	vs = append([]interface{}{i.jenkins.Header}, vs...)
	resp, err := i.jenkins.Req.Do(method, i.Url+entry, vs...)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode >= 400 {
		return nil, fmt.Errorf(resp.Dump())
	}
	return resp, nil
}

func (i *Item) APIJson(param req.Param) (gjson.Result, error) {
	resp, err := i.Request("GET", "api/json", param)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(resp.String()), nil
}

func (i *Item) String() string {
	return fmt.Sprintf("<%s: %s>", i.Class, i.Url)
}

func (i *Item) GetClass() string {
	return i.Class
}

func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	return reg.Split(text, -1)
}

func GetClassName(text string) string {
	ss := RegSplit(text, "[.$]")
	return ss[len(ss)-1]
}

func AppendSlash(url string) string {
	if strings.HasSuffix(url, "/") {
		return url
	}
	return url + "/"
}

func getId(url string) int {
	_, base := path.Split(strings.Trim(url, "/"))
	id, _ := strconv.Atoi(base)
	return id
}
