package raft

type Server struct {
	ID       string
	Address  string
	Suffrage string
	Leader   bool
}

type Stats map[string]string

type Info struct {
	State   string
	Addr    string
	ID      string
	Servers []Server
	Stats   Stats
}
