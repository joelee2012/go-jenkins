package jenkins

import (
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

func (v *ViewService) Get(name string) (*View, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(ReqParams{"tree": "views[name,url,description]"}, jobJson); err != nil {
		return nil, err
	}
	for _, view := range jobJson.Views {
		if view.Name == name {
			return view, nil
		}
	}
	return nil, nil
}

func (v *ViewService) Create(name, xml string) error {
	_, err := v.Request("POST", "createView", ReqParams{"name": name}, req.BodyXML(xml))
	return err
}
func (v *ViewService) Delete(name string) error {
	_, err := v.Request("POST", "view/"+name+"/doDelete")
	return err
}

func (v *ViewService) AddJobToView(name, job string) error {
	_, err := v.Request("POST", "view/"+name+"/addJobToView", ReqParams{"name": job})
	return err
}

func (v *ViewService) RemoveJobFromView(name, job string) error {
	_, err := v.Request("POST", "view/"+name+"/removeJobFromView", ReqParams{"name": job})
	return err
}

func (v *ViewService) GetConfigure(name string) (string, error) {
	resp, err := v.Request("GET", "view/"+name+"/config.xml")
	return resp.String(), err
}

func (v *ViewService) SetConfigure(name, xml string) error {
	_, err := v.Request("POST", "view/"+name+"/config.xml", req.BodyXML(xml))
	return err
}

func (v *ViewService) SetDescription(name, description string) error {
	_, err := v.Request("POST", "view/"+name+"/submitDescription", ReqParams{"description": description})
	return err
}

func (v *ViewService) List() ([]*View, error) {
	jobJson := &Job{}
	if err := v.BindAPIJson(ReqParams{"tree": "views[name,url,description]"}, jobJson); err != nil {
		return nil, err
	}
	return jobJson.Views, nil
}

func (v *ViewService) bindViewAPIJson(name string, o interface{}) error {
	resp, err := v.Request("GET", "view/"+name+"/api/json")
	if err != nil {
		return err
	}
	return resp.ToJSON(o)
}

func (v *ViewService) GetJobFromView(name, job string) (*JobItem, error) {
	view := &View{}
	if err := v.bindViewAPIJson(name, view); err != nil {
		return nil, err
	}
	for _, job := range view.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, v.client), nil
		}
	}
	return nil, nil
}

func (v *ViewService) ListJobInView(name string) ([]*JobItem, error) {
	view := &View{}
	if err := v.bindViewAPIJson(name, view); err != nil {
		return nil, err
	}
	var jobs []*JobItem
	for _, job := range view.Jobs {
		jobs = append(jobs, NewJobItem(job.URL, job.Class, v.client))
	}
	return jobs, nil
}
