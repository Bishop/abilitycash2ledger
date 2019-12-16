package main

type Scope struct {
	Datafiles []*Datafile `json:"datafile"`
}

type Datafile struct {
	Path     string            `json:"path"`
	Target   string            `json:"target"`
	Accounts map[string]string `json:"accounts"`
}
