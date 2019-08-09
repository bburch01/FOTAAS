// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {

	var fs http.FileSystem = http.Dir("./assets")

	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName: "assets",
		VariableName: "Assets",
		Filename: "./generated/assets/assets_vfsdata.go",
	})
	if err != nil {
		log.Fatalln(err)
	}

}
