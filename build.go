package jenkins

import (
	"bufio"
	"regexp"
)

type BuildItem struct {
	*Jenkins
	BaseURL string
	Class   string
	ID      int
}

func NewBuildItem(url, class string, jenkins *Jenkins) *BuildItem {
	return &BuildItem{Jenkins: jenkins, BaseURL: url, Class: class, ID: parseId(url)}
}

func (b *BuildItem) LoopLog(f func(line string) error) error {
	resp, err := R().Get(b.BaseURL + "consoleText")
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
	_, err := R().SetQueryParam("tree", "building").SetSuccessResult(build).Get(b.BaseURL + "api/json")
	return build.Building, err
}

func (b *BuildItem) GetResult() (string, error) {
	status := make(map[string]string)
	_, err := R().SetQueryParam("tree", "result").SetSuccessResult(&status).Get(b.BaseURL + "api/json")
	return status["result"], err
}

func (b *BuildItem) Delete() error {
	_, err := R().Post(b.BaseURL + "doDelete")
	return err
}

func (b *BuildItem) Stop() error {
	_, err := R().Post(b.BaseURL + "stop")
	return err
}

func (b *BuildItem) Kill() error {
	_, err := R().Post(b.BaseURL + "kill")
	return err
}

func (b *BuildItem) Term() error {
	_, err := R().Post(b.BaseURL + "term")
	return err
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *BuildItem) GetJob() (*JobItem, error) {
	jobName, _ := b.URL2Name(re.ReplaceAllLiteralString(b.BaseURL, ""))
	return b.Jenkins.GetJob(jobName)
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
		resp, err := R().SetPathParam("start", start).Get(b.BaseURL + entry)
		if err != nil {
			return err
		}
		if start == resp.Header.Get("X-Text-Size") {
			continue
		}
		if err := f(resp.String()); err != nil {
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
	if _, err := R().SetQueryParam("tree", "description").SetSuccessResult(&data).Get(b.BaseURL + "api/json"); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (b *BuildItem) SetDescription(description string) error {
	_, err := R().SetQueryParam("description", description).Post(b.BaseURL + "submitDescription")
	return err
}
