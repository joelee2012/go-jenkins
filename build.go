package jenkins

import "regexp"

type Build struct {
	Item
	ID int
}

type BuildShortJson struct {
	Class    string `json:"_class"`
	Number   int    `json:"number"`
	URL      string `json:"url"`
	Building bool   `json:"building"`
}

func NewBuild(url, class string, jenkins *Jenkins) *Build {
	return &Build{Item: *NewItem(url, class, jenkins), ID: parseId(url)}
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

func (b *Build) Delete() error {
	return doDelete(b)
}

func (b *Build) Stop() error {
	_, err := b.Request("POST", "stop")
	return err
}

func (b *Build) Kill() error {
	_, err := b.Request("POST", "kill")
	return err
}

func (b *Build) Term() error {
	_, err := b.Request("POST", "term")
	return err
}

func (b *Build) GetJob() (*Job, error) {
	re := regexp.MustCompile(`\w+[/]?$`)
	jobName, _ := b.jenkins.URLToName(re.ReplaceAllLiteralString(b.URL, ""))
	return b.jenkins.GetJob(jobName)
}
