package main

import (
	"github.com/weastur/maf/cmd"
	mysql "github.com/weastur/maf/pkg"
)

func main() {
	mysql.Foo()
	cmd.Execute()
}
