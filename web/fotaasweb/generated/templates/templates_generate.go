// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {

	var fs http.FileSystem = http.Dir("./templates")

	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "templates",
		VariableName: "Templates",
		Filename:     "./generated/templates/templates_vfsdata.go",
	})
	if err != nil {
		log.Fatalln(err)
	}

}
