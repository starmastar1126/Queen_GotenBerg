// Package converter implements a solution for converting one or more files to PDF.
package converter

import (
	"fmt"
	"io"
	"net/http"
	"os"

	gfile "github.com/gulien/gotenberg/app/converter/file"
	"github.com/gulien/gotenberg/app/converter/process"
	ghttp "github.com/gulien/gotenberg/app/http"

	"github.com/satori/go.uuid"
)

// Converter handles conversion into PDF of files coming from a request.
type Converter struct {
	files      []*gfile.File
	workingDir string
}

// NewConverter instantiates a converter by parsing a request.
func NewConverter(r *http.Request, contentType ghttp.ContentType) (*Converter, error) {
	c := &Converter{
		workingDir: fmt.Sprintf("./%s/", uuid.NewV4().String()),
	}

	if err := os.Mkdir(c.workingDir, 0666); err != nil {
		return nil, err
	}

	switch contentType {
	case ghttp.MultipartFormDataContentType:
		reader, err := r.MultipartReader()
		if err != nil {
			return nil, err
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			if part.FileName() == "" {
				continue
			}

			f, err := gfile.NewFile(c.workingDir, part)
			if err != nil {
				return nil, err
			}

			c.files = append(c.files, f)
		}
		break
	default:
		f, err := gfile.NewFile(c.workingDir, r.Body)
		if err != nil {
			return nil, err
		}

		c.files = append(c.files, f)
	}

	return c, nil
}

// Convert converts its associated files to PDF. If more than one file,
// it will merge all of them into one unique PDF file.
// Returns the new file path or an error if something bad happened.
func (c *Converter) Convert() (string, error) {
	var filesPaths []string
	for _, f := range c.files {
		if f.Type != gfile.PDFType {
			path, err := process.Unconv(c.workingDir, f)
			if err != nil {
				return "", err
			}

			filesPaths = append(filesPaths, path)
		} else {
			filesPaths = append(filesPaths, f.Path)
		}
	}

	if len(filesPaths) == 1 {
		return filesPaths[0], nil
	}

	path, err := process.Merge(c.workingDir, filesPaths)
	if err != nil {
		return "", err
	}

	return path, nil
}

// Clear removes all file inside its working directory.
func (c *Converter) Clear() error {
	return os.RemoveAll(c.workingDir)
}
