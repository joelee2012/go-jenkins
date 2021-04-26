package jenkins

import "github.com/imroc/req"

type ClassReader interface {
	GetClass() string
}

type Deleter interface {
	Delete() error
}

type ConfigSetter interface {
	SetConfigure(xml string) error
}

type ConfigGetter interface {
	GetConfigure() (string, error)
}

type Configurator interface {
	ConfigGetter
	ConfigSetter
}

type JobInterface interface {
	Rename(name string) error
	Move(path string) error
	IsBuildable() (bool, error)
	GetName() string
	GetFullName() string
	GetFullDisplayName() string
	GetParent() (Job, error)
	Deleter
	Configurator
	ClassReader
}

type Requester interface {
	Request(method, entry string, vs ...interface{}) (*req.Resp, error)
}
