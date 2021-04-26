package jenkins

import (
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

type Build interface {
	GetConsoleText() (string, error)
	IsBuilding() bool
	APIJson(param req.Param) (gjson.Result, error)
}

type BaseBuild struct {
	Item
}

func NewBuild(url, class string, jenkins *Jenkins) Build {
	build := BaseBuild{
		Item: Item{
			Url:     url,
			Class:   class,
			jenkins: jenkins,
		},
	}

	switch class {
	case "WorkflowRun":
		return &WorkflowRun{BaseBuild: build}
	case "FreeStyleBuild":
		return &FreeStyleBuild{BaseBuild: build}
	case "MatrixBuild":
		return &MatrixBuild{BaseBuild: build}
	default:
		return &build
	}
}

func (b *BaseBuild) GetConsoleText() (string, error) {
	resp, err := b.Request("GET", "consoleText", req.Param{})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (b *BaseBuild) IsBuilding() bool {
	result, err := b.APIJson(req.Param{"tree": "building"})
	if err != nil {
		panic(err)
	}
	return result.Get("building").Bool()
}

func (b *BaseBuild) GetNumber() int {
	return getId(b.Url)
}

type WorkflowRun struct {
	BaseBuild
}

type FreeStyleBuild struct {
	BaseBuild
}
type MatrixBuild struct {
	BaseBuild
}
