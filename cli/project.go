package cli

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

// Project contains name, license and paths to projects.
type Project struct {
	PkgName      string
	Copyright    string
	DockerName   string
	AbsolutePath string
	AppName      string
	TemplatePath string
	UrlBase      string
}

func (p *Project) Create() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(p.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.Mkdir(p.AbsolutePath, 0754); err != nil {
			return err
		}
	}

	return filepath.Walk(p.TemplatePath, p.walkHandler)
}

func (p *Project) walkHandler(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Errorf("error in filepath walk handler: %s", err)
		return err
	}

	log.Debugf("path is %s, teplate path is %s", path, p.TemplatePath)

	log.Infof("looking at file %s: dir: %t", path, info.IsDir())

	if path == p.TemplatePath {
		log.Debugf("skipping the base directory: %s", path)
		return nil
	}

	if info.IsDir() && path == p.TemplatePath+"/.git" {
		log.Debugf("skipping the git directory: %s", path)
		return nil
	} else if strings.HasPrefix(path, p.TemplatePath+"/.git/") {
		log.Debugf("skipping the git directory contents: %s", path)
		return nil
	}

	if info.IsDir() {
		newdir := filepath.Clean(p.AbsolutePath + "/" + strings.TrimPrefix(path, p.TemplatePath))
		log.Debugf("making the directory: %s", newdir)
		return os.MkdirAll(newdir, 0754)
	}

	newfile := filepath.Clean(p.AbsolutePath + "/" + strings.TrimPrefix(path, p.TemplatePath))
	if strings.HasSuffix(path, ".tmpl") {
		log.Debugf("%s is a template file", path)

		newfile = strings.TrimSuffix(newfile, ".tmpl")

		blob, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New("config").Parse(string(blob))
		if err != nil {
			return err
		}

		var b bytes.Buffer
		parsedTemplate := bufio.NewWriter(&b)
		err = tmpl.Execute(parsedTemplate, p)
		if err != nil {
			return err
		}
		parsedTemplate.Flush()

		log.Debugf("writing output to %s", newfile)
		if err := ioutil.WriteFile(newfile, b.Bytes(), 0644); err != nil {
			return err
		}

		return nil
	}

	log.Debugf("%s is a regular file", path)

	if _, err := os.Stat(newfile); !os.IsNotExist(err) {
		log.Infof("file %s already exists, not overwriting", newfile)
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(newfile, data, 0644)
}
