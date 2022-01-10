package jenkins

import (
	"fmt"

	"github.com/imroc/req"
)

func doRequest(j *Jenkins, method, url string, vs ...interface{}) (*req.Resp, error) {
	if _, err := j.GetCrumb(); err != nil {
		return nil, err
	}
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

func doRequestAndDropResp(r Requester, method, entry string, vs ...interface{}) error {
	_, err := r.Request(method, entry, vs...)
	return err
}

func doDelete(r Requester) error {
	return doRequestAndDropResp(r, "POST", "doDelete")
}

func doGetConfigure(r Requester) (string, error) {
	resp, err := r.Request("GET", "config.xml")
	return resp.String(), err
}

func doSetConfigure(r Requester, xml string) error {
	return doRequestAndDropResp(r, "POST", "config.xml", req.BodyXML(xml))
}

func doSetDescription(r Requester, description string) error {
	return doRequestAndDropResp(r, "POST", "submitDescription", ReqParams{"description": description})
}

func doGetDescription(r Requester) (string, error) {
	data := make(map[string]string)
	if err := doBindAPIJson(r, ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func doBindAPIJson(r Requester, param ReqParams, v interface{}) error {
	resp, err := r.Request("GET", "api/json", param)
	if err != nil {
		return err
	}
	return resp.ToJSON(v)
}

func doDisable(r Requester) error {
	return doRequestAndDropResp(r, "POST", "disable")
}

func doEnable(r Requester) error {
	return doRequestAndDropResp(r, "POST", "enable")
}
