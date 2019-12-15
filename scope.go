package main

type Scope struct {
	Datafiles []*Datafile `json:"datafile"`
}

type Datafile struct {
	Path               string   `json:"path"`
	ConsiderLockedFlag bool     `json:"consider_locked_flag"`
	Accounts           []string `json:"accounts"`
}
