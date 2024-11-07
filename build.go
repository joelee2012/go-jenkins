package jenkins

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type Build struct {
	*Item
	Number int
}

func NewBuild(url, class string, jenkins *Jenkins) *Build {
	return &Build{Item: NewItem(url, class, jenkins), Number: parseId(url)}
}

func (b *Build) IsBuilding() (bool, error) {
	var build struct {
		Class    string `json:"_class"`
		Building bool   `json:"building"`
	}
	err := b.ApiJson(&build, &ApiJsonOpts{Tree: "building"})
	return build.Building, err
}

func (b *Build) GetResult() (string, error) {
	status := make(map[string]string)
	err := b.ApiJson(&status, &ApiJsonOpts{Tree: "result"})
	return status["result"], err
}

func (b *Build) Delete() (*http.Response, error) {
	return b.Request("POST", "doDelete", nil)
}

func (b *Build) Stop() (*http.Response, error) {
	return b.Request("POST", "stop", nil)
}

func (b *Build) Kill() (*http.Response, error) {
	return b.Request("POST", "kill", nil)
}

func (b *Build) Term() (*http.Response, error) {
	return b.Request("POST", "term", nil)
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *Build) GetJob() (*Job, error) {
	jobName, _ := b.jenkins.URL2Name(re.ReplaceAllLiteralString(b.URL, ""))
	return b.jenkins.GetJob(jobName)
}

func (b *Build) LoopLog(f func(line string) error) error {
	resp, err := b.Request("GET", "consoleText", nil)
	if err != nil {
		return err
	}
	return scanResponse(resp, f)
}

func scanResponse(resp *http.Response, f func(line string) error) error {
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if err := f(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (b *Build) LoopProgressiveLog(kind string, f func(line string) error) error {
	var entry string
	switch kind {
	case "html":
		entry = "logText/progressiveHtml"
	case "text":
		entry = "logText/progressiveText"
	default:
		panic("kind must be html or text")
	}
	start := "0"
	for {
		resp, err := b.Request("GET", fmt.Sprintf("%s?start=%s", entry, start), nil)
		if err != nil {
			return err
		}
		if start == resp.Header.Get("X-Text-Size") {
			continue
		}
		if err := scanResponse(resp, f); err != nil {
			return err
		}
		if resp.Header.Get("X-More-Data") != "true" {
			break
		}
		start = resp.Header.Get("X-Text-Size")
	}
	return nil
}

func (b *Build) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := b.ApiJson(&data, &ApiJsonOpts{Tree: "description"}); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (b *Build) SetDescription(description string) (*http.Response, error) {
	v := url.Values{}
	v.Add("description", description)
	return b.Request("POST", "submitDescription?"+v.Encode(), nil)
}
