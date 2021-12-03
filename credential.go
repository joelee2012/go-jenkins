package jenkins

import "github.com/imroc/req"

type Credentials struct {
	Item
}

func (cs *Credentials) Get(name string) (*Credential, error) {
	var credsJson CredentialsJson
	if err := cs.BindAPIJson(ReqParams{"depth": "1"}, &credsJson); err != nil {
		return nil, err
	}
	for _, cred := range credsJson.Credentials {
		if cred.ID == name {
			return &Credential{Item: *NewItem(cs.URL+"credential/"+name, "Credential", cs.jenkins)}, nil
		}
	}
	return nil, nil
}

func (cs *Credentials) Create(xml string) error {
	_, err := cs.Request("POST", "createCredentials", req.BodyXML(xml))
	return err
}

func (cs *Credentials) Delete(name string) error {
	cred, err := cs.Get(name)
	if err != nil {
		return err
	}
	return cred.Delete()
}

func (cs *Credentials) List() ([]*Credential, error) {
	var credsJson CredentialsJson
	if err := cs.BindAPIJson(ReqParams{"depth": "1"}, &credsJson); err != nil {
		return nil, err
	}
	var creds []*Credential
	for _, cred := range credsJson.Credentials {
		creds = append(creds, &Credential{Item: *NewItem(cs.URL+"credential/"+cred.ID, "Credential", cs.jenkins)})
	}
	return creds, nil
}

type Credential struct {
	Item
}

func (c *Credential) Delete() error {
	return doDelete(c)
}

func (c *Credential) SetConfigure(xml string) error {
	return doSetConfigure(c, xml)
}

func (c *Credential) GetConfigure(xml string) (string, error) {
	return doGetConfigure(c)
}
