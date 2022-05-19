package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"text/template"
)

// note the all: option in go 1.18 will load hidden files so we dont have to include `cp` instructions in readme for pre-commit.
//go:embed templates/* templates/terraform/.pre-commit-config.yaml templates/terraform/.gitignore
var templatesFs embed.FS

type TemplateData struct {
	Name        string
	Description string
	Access      string
	Type        string
	OutputDir   string
}

func Generate(data *TemplateData) error {
	templateFiles, _ := fs.Sub(fs.FS(templatesFs), "templates/terraform")

	err := fs.WalkDir(templateFiles, ".", func(path string, info fs.DirEntry, err error) error {
		outputPath := fmt.Sprintf("%s/%s", data.OutputDir, path)
		if info.IsDir() {
			if path == "." {
				return os.MkdirAll(data.OutputDir, 0777)
			}

			return os.Mkdir(outputPath, 0777)
		}

		var tmpl *template.Template
		var outputFile *os.File
		tmpl, _ = template.ParseFS(templateFiles, path)
		outputFile, err = os.Create(outputPath)

		if err != nil {
			return err
		}

		defer outputFile.Close()
		return tmpl.Execute(outputFile, data)
	})

	return err
}
