package reify

import (
	"bytes"
	"html/template"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

// Reify returns the resulting JSON by expanding the template using the
// supplied data.
func Reify(templateFileName string, templateData interface{}) (json []byte, err error) {
	template, err := template.ParseFiles(templateFileName)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, templateData)
	if err != nil {
		return nil, err
	}

	// Translate YAML to JSON.
	json, err = yaml.YAMLToJSON(buf.Bytes())

	glog.Infof("reified template [%s] with data [%v]:\nYAML:\n%s\n\nJSON:\n%s", templateFileName, templateData, buf.String(), string(json))

	return
}
