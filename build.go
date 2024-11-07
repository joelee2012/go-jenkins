package jenkins

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type BuildItem struct {
	*Item
	Number int
}

func NewBuildItem(url, class string, jenkins *Jenkins) *BuildItem {
	return &BuildItem{Item: NewItem(url, class, jenkins), Number: parseId(url)}
}

func (b *BuildItem) LoopLog(f func(line string) error) error {
	resp, err := b.Request("GET", "consoleText", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if err := f(scanner.Text()); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (b *BuildItem) IsBuilding() (bool, error) {
	var build struct {
		Class    string `json:"_class"`
		Building bool   `json:"building"`
	}
	err := b.ApiJson(&build, &ApiJsonOpts{Tree: "building"})
	return build.Building, err
}

func (b *BuildItem) GetResult() (string, error) {
	status := make(map[string]string)
	err := b.ApiJson(&status, &ApiJsonOpts{Tree: "result"})
	return status["result"], err
}

func (b *BuildItem) Delete() (*http.Response, error) {
	return b.Request("POST", "doDelete", nil)
}

func (b *BuildItem) Stop() (*http.Response, error) {
	return b.Request("POST", "stop", nil)
}

func (b *BuildItem) Kill() (*http.Response, error) {
	return b.Request("POST", "kill", nil)
}

func (b *BuildItem) Term() (*http.Response, error) {
	return b.Request("POST", "term", nil)
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *BuildItem) GetJob() (*JobItem, error) {
	jobName, _ := b.jenkins.URL2Name(re.ReplaceAllLiteralString(b.URL, ""))
	return b.jenkins.GetJob(jobName)
}

func (b *BuildItem) LoopProgressiveLog(kind string, f func(data []byte) error) error {
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
		resp, err := b.Request("GET", entry+"?start="+start, nil)
		if err != nil {
			return err
		}

		if start == resp.Header.Get("X-Text-Size") {
			continue
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if err := f(data); err != nil {
			return err
		}
		if resp.Header.Get("X-More-Data") != "true" {
			break
		}
		start = resp.Header.Get("X-Text-Size")
	}
	return nil
}

func (b *BuildItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := b.ApiJson(&data, &ApiJsonOpts{Tree: "description"}); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (b *BuildItem) SetDescription(description string) (*http.Response, error) {
	v := url.Values{}
	v.Add("description", description)
	return b.Request("POST", "submitDescription?"+v.Encode(), nil)
}
