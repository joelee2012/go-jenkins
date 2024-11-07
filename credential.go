package jenkins

import (
	"fmt"
	"io"
	"net/http"
)

type Credentials struct {
	*Item
}

func (cs *Credentials) Get(name string) (*CredentialJson, error) {
	var credsJson CredentialsJson
	if err := cs.ApiJson(&credsJson, &ApiJsonOpts{Depth: 1}); err != nil {
		return nil, err
	}
	if credsJson.Credentials != nil {
		for _, cred := range credsJson.Credentials {
			if cred.ID == name {
				return cred, nil
			}
		}
	}
	return nil, fmt.Errorf("no such credential [%s]", name)
}

func (cs *Credentials) Create(xml io.Reader) (*http.Response, error) {
	return cs.Request("POST", "createCredentials", xml)
}

func (cs *Credentials) Delete(name string) (*http.Response, error) {
	return cs.Request("POST", "credential/"+name+"/doDelete", nil)
}

func (cs *Credentials) GetConfigure(name string) (string, error) {
	return readResponseToString(cs, "GET", "credential/"+name+"/config.xml", nil)
}

func (cs *Credentials) SetConfigure(name string, xml io.Reader) (*http.Response, error) {
	return cs.Request("POST", "credential/"+name+"/config.xml", xml)
}

func (cs *Credentials) List() ([]*CredentialJson, error) {
	var credsJson CredentialsJson
	if err := cs.ApiJson(&credsJson, &ApiJsonOpts{Depth: 1}); err != nil {
		return nil, err
	}
	return credsJson.Credentials, nil
}
