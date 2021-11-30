package jenkins

import (
	"github.com/imroc/req"
)

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
	resp, err := b.Request("GET", "consoleText", req.Param{})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (b *Build) IsBuilding() (bool, error) {
	var status BuildShortJson
	err := b.BindAPIJson(req.Param{"tree": "building"}, &status)
	return status.Building, err
}

