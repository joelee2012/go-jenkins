package jenkins

import (
	"fmt"
	"log"

	"github.com/imroc/req"
)

type Deleter interface {
	Delete() error
}

type Configurator interface {
	GetConfigure() (string, error)
}

type Job interface {
	Rename(name string) error
	Move(path string)
	Deleter
	Configurator
}

type BaseJob struct {
	Item
}

func NewJob(url, class string, jenkins *Jenkins) Job {
	job := BaseJob{
		Item: Item{
			Url:     url,
			Class:   class,
			jenkins: jenkins},
	}
	switch class {
	case "WorkflowJob":
		return &WorkflowJob{Project: Project{BaseJob: job}}
	case "Folder":
		return &Folder{BaseJob: job}
	default:
		return &job
	}
}

func (j *BaseJob) Rename(name string) error {
	log.Println("rename")
	resp, err := j.Request("POST", "confirmRename", req.Param{"newName": name})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.Url = url.String()
	return nil
}

func (j *BaseJob) Move(path string) {
	log.Println("move")
}

func (j *BaseJob) Delete() error {
	_, err := j.Request("POST", "doDelete")
	return err
}

func (j *BaseJob) GetParent() {
	log.Println("get parent")
}

func (j *BaseJob) GetConfigure() (string, error) {
	resp, err := j.Request("GET", "config.xml")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

type Project struct {
	BaseJob
}

func (p *Project) Build(param req.Param) (*QueueItem, error) {
	resp, err := p.Request("POST", "build", param)
	if err != nil {
		return nil, err
	}
	url, err := resp.Response().Location()
	if err != nil {
		return nil, err
	}
	return NewQueueItem(url.String(), p.jenkins), nil
}

type WorkflowJob struct {
	Project
}

type Folder struct {
	BaseJob
}

func (f *Folder) GetJob(name string) (Job, error) {
	apiJson, err := f.APIJson(req.Param{"tree": "jobs[url,name]"})
	if err != nil {
		return nil, err
	}
	result := apiJson.Get(fmt.Sprintf(`jobs.#(name=="%s")`, name))
	if !result.Exists() {
		return nil, fmt.Errorf("no such Job[%s]", name)
	}
	url := result.Get("url").String()
	class := GetClassName(result.Get("_class").String())
	return NewJob(url, class, f.jenkins), nil
}

func (f *Folder) CreateJob(name, xml string) error {
	_, err := f.Request("POST", "createItem", req.Param{"name": name}, req.BodyXML(xml))
	return err
}
