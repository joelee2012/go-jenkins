package jenkins

import (
	"regexp"
)

type Build struct {
	*Item
	ID int
}

type BuildShortJson struct {
	Class    string `json:"_class"`
	Number   int    `json:"number"`
	URL      string `json:"url"`
	Building bool   `json:"building"`
}

func NewBuild(url, class string, jenkins *Jenkins) *Build {
	return &Build{Item: NewItem(url, class, jenkins), ID: parseId(url)}
}

func (b *Build) GetConsoleText() (string, error) {
	resp, err := b.Request("GET", "consoleText", ReqParams{})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (b *Build) IsBuilding() (bool, error) {
	var status BuildShortJson
	err := b.BindAPIJson(ReqParams{"tree": "building"}, &status)
	return status.Building, err
}

func (b *Build) GetResult() (string, error) {
	var status map[string]string
	err := b.BindAPIJson(ReqParams{"tree": "result"}, &status)
	return status["result"], err
}

func (b *Build) Delete() error {
	return doDelete(b)
}

func (b *Build) Stop() error {
	return doRequestAndDropResp(b, "POST", "stop")
}

func (b *Build) Kill() error {
	return doRequestAndDropResp(b, "POST", "kill")
}

func (b *Build) Term() error {
	return doRequestAndDropResp(b, "POST", "term")
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *Build) GetJob() (*Job, error) {
	jobName, _ := b.jenkins.URLToName(re.ReplaceAllLiteralString(b.URL, ""))
	return b.jenkins.GetJob(jobName)
}

func (b *Build) IterateProgressConsoleText(kind string, f func(line string) error) error {
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
		if resp.Response().Header.Get("X-More-Data") != "True" {
			break
		}
		start = resp.Response().Header.Get("X-Text-Size")
	}
	return nil
}
