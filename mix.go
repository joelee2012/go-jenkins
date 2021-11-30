package jenkins

import (
	"fmt"

	"github.com/imroc/req"
)

func doRequest(j *Jenkins, method, url string, vs ...interface{}) (*req.Resp, error) {
	vs = append(vs, j.Header)
	resp, err := j.Req.Do(method, url, vs...)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode >= 400 {
		return nil, fmt.Errorf(resp.Dump())
	}
	return resp, nil
}

func doDelete(r Requester) error {
	_, err := r.Request("POST", "doDelete")
	return err
}

func doGetConfigure(r Requester) (string, error) {
	resp, err := r.Request("GET", "config.xml")
	return resp.String(), err
}

func doSetConfigure(r Requester, xml string) error {
	_, err := r.Request("POST", "config.xml", req.BodyXML(xml))
	return err
}

func doSetDescription(r Requester, description string) error {
	_, err := r.Request("POST", "submitDescription", description)
	return err
}

func doGetDescription(r Requester) (string, error) {
	data := make(map[string]string)
	if err := doBindAPIJson(r, req.Param{"tree": "description"}, data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func doBindAPIJson(r Requester, param req.Param, v interface{}) error {
	resp, err := r.Request("GET", "api/json", param)
	if err != nil {
		return err
	}
	return resp.ToJSON(v)
}
