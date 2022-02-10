package jenkins

import (
	"strings"

	"github.com/imroc/req"
)

type ViewService struct {
	*Item
}

func NewViewService(v interface{}) *ViewService {
	if c, ok := v.(*Client); ok {
		return &ViewService{Item: NewItem(c.URL, "Views", c)}
	}
	if c, ok := v.(*JobItem); ok {
		return &ViewService{Item: NewItem(c.URL, "Views", c.client)}
	}
	return nil
}

func (v *ViewService) Get(name string) (*ViewItem, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(ReqParams{"tree": "views[name,url]"}, jobJson); err != nil {
		return nil, err
	}
	for _, view := range jobJson.Views {
		if view.Name == name {
			return NewViewItem(view, v.client), nil
		}
	}
	return nil, nil
}

func (v *ViewService) Create(name, xml string) error {
	_, err := v.Request("POST", "createView", ReqParams{"name": name}, req.BodyXML(xml))
	return err
}

func (v *ViewService) List() ([]*ViewItem, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(ReqParams{"tree": "views[name,url]"}, jobJson); err != nil {
		return nil, err
	}
	var views []*ViewItem
	for _, view := range jobJson.Views {
		views = append(views, NewViewItem(view, v.client))
	}
	return views, nil
}

type ViewItem struct {
	*Item
	Name string
}

func NewViewItem(view *View, client *Client) *ViewItem {
	// name of all view for jenkins is 'all', but for folder is 'All'
	url := view.URL
	if strings.HasSuffix(view.Class, "AllView") {
		if client.URL == view.URL {
			url += "view/all"
		}
		url += "view/All"
	}
	return &ViewItem{Item: NewItem(url, view.Class, client), Name: view.Name}
}

func (v *ViewItem) GetJob(name string) (*JobItem, error) {
	view := &View{}
	if err := v.BindAPIJson(ReqParams{"tree": "jobs[name,url]"}, view); err != nil {
		return nil, err
	}
	for _, job := range view.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, v.client), nil
		}
	}
	return nil, nil
}

func (v *ViewItem) IncludeJob(name string) error {
	_, err := v.Request("POST", "addJobToView", ReqParams{"name": name})
	return err
}

func (v *ViewItem) ExcludeJob(name string) error {
	_, err := v.Request("POST", "removeJobFromView", ReqParams{"name": name})
	return err
}

func (v *ViewItem) GetConfigure() (string, error) {
	resp, err := v.Request("GET", "/config.xml")
	return resp.String(), err
}

func (v *ViewItem) SetConfigure(xml string) error {
	_, err := v.Request("POST", "/config.xml", req.BodyXML(xml))
	return err
}

func (v *ViewItem) Delete() error {
	_, err := v.Request("POST", "doDelete")
	return err
}

func (v *ViewItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := v.BindAPIJson(ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (v *ViewItem) SetDescription(description string) error {
	_, err := v.Request("POST", "submitDescription", ReqParams{"description": description})
	return err
}

func (v *ViewItem) ListJob() ([]*JobItem, error) {
	view := &View{}
	if err := v.BindAPIJson(ReqParams{"tree": "jobs[name,url]"}, view); err != nil {
		return nil, err
	}
	var jobs []*JobItem
	for _, job := range view.Jobs {
		jobs = append(jobs, NewJobItem(job.URL, job.Class, v.client))
	}
	return jobs, nil
}
