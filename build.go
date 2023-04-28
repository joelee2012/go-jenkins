package jenkins

import (
	"bufio"
	"regexp"
)

type BuildItem struct {
	*Client
	BaseURL string
	Class   string
	ID      int
}

func NewBuildItem(url, class string, client *Client) *BuildItem {
	return &BuildItem{Client: client, BaseURL: url, Class: class, ID: parseId(url)}
}

func (b *BuildItem) LoopLog(f func(line string) error) error {
	resp, err := b.R().Get("consoleText")
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

func (b *BuildItem) BindAPIJson(params map[string]string, v interface{}) error {
	_, err := b.R().SetQueryParams(params).SetSuccessResult(v).Get("api/json")
	return err
}

func (b *BuildItem) IsBuilding() (bool, error) {
	var build struct {
		Class    string `json:"_class"`
		Building bool   `json:"building"`
	}
	err := b.BindAPIJson(map[string]string{"tree": "building"}, &build)
	return build.Building, err
}

func (b *BuildItem) GetResult() (string, error) {
	status := make(map[string]string)
	err := b.BindAPIJson(map[string]string{"tree": "result"}, &status)
	return status["result"], err
}

func (b *BuildItem) Delete() error {
	_, err := b.R().Post("doDelete")
	return err
}

func (b *BuildItem) Stop() error {
	_, err := b.R().Post("stop")
	return err
}

func (b *BuildItem) Kill() error {
	_, err := b.R().Post("kill")
	return err
}

func (b *BuildItem) Term() error {
	_, err := b.R().Post("term")
	return err
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *BuildItem) GetJob() (*JobItem, error) {
	jobName, _ := b.client.URL2Name(re.ReplaceAllLiteralString(b.BaseURL, ""))
	return b.client.GetJob(jobName)
}

func (b *BuildItem) LoopProgressiveLog(kind string, f func(line string) error) error {
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
		resp, err := b.Request("GET", entry, ReqParams{"start": start})
		if err != nil {
			return err
		}
		if start == resp.Response().Header.Get("X-Text-Size") {
			continue
		}
		if err := f(resp.String()); err != nil {
			return err
		}
		if resp.Response().Header.Get("X-More-Data") != "true" {
			break
		}
		start = resp.Response().Header.Get("X-Text-Size")
	}
	return nil
}

func (b *BuildItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := b.BindAPIJson(ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (b *BuildItem) SetDescription(description string) error {
	_, err := b.Request("POST", "submitDescription", ReqParams{"description": description})
	return err
}
