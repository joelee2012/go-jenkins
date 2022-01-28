package jenkins

import (
	"bufio"
	"regexp"
)

type BuildService struct {
	*Item
	ID int
}

func NewBuild(url, class string, client *Client) *BuildService {
	return &BuildService{Item: NewItem(url, class, client), ID: parseId(url)}
}

func (b *BuildService) LoopLog(f func(line string) error) error {
	resp, err := b.Request("GET", "consoleText", ReqParams{})
	if err != nil {
		return err
	}
	defer resp.Response().Body.Close()
	scanner := bufio.NewScanner(resp.Response().Body)
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

func (b *BuildService) IsBuilding() (bool, error) {
	var build struct {
		Class    string `json:"_class"`
		Building bool   `json:"building"`
	}
	err := b.BindAPIJson(ReqParams{"tree": "building"}, &build)
	return build.Building, err
}

func (b *BuildService) GetResult() (string, error) {
	status := make(map[string]string)
	err := b.BindAPIJson(ReqParams{"tree": "result"}, &status)
	return status["result"], err
}

func (b *BuildService) Delete() error {
	_, err := b.Request("POST", "doDelete")
	return err
}

func (b *BuildService) Stop() error {
	_, err := b.Request("POST", "stop")
	return err
}

func (b *BuildService) Kill() error {
	_, err := b.Request("POST", "kill")
	return err
}

func (b *BuildService) Term() error {
	_, err := b.Request("POST", "term")
	return err
}

var re = regexp.MustCompile(`\w+[/]?$`)

func (b *BuildService) GetJob() (*JobService, error) {
	jobName, _ := b.client.URLToName(re.ReplaceAllLiteralString(b.URL, ""))
	return b.client.GetJob(jobName)
}

func (b *BuildService) LoopProgressiveLog(kind string, f func(line string) error) error {
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

func (b *BuildService) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := b.BindAPIJson(ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (b *BuildService) SetDescription(description string) error {
	_, err := b.Request("POST", "submitDescription", ReqParams{"description": description})
	return err
}
