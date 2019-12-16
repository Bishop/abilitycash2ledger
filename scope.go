package main

type Scope struct {
	Datafiles []*Datafile `json:"datafile"`
}

type Datafile struct {
	Path     string   `json:"path"`
	Target   string   `json:"target"`
	Accounts []string `json:"accounts"`
}
