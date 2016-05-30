package load

import (
	"bytes"
	"encoding/json"
	"runtime"
	"text/template"

	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	load "github.com/shirou/gopsutil/load"
)

var log = logging.MustGetLogger("go3status.load")

type LoadInstance struct {
	name     string
	format   string
	template *template.Template
}

func (t LoadInstance) RefreshInterval() int {
	return 5
}

func (t LoadInstance) Name() (n string) {
	n = t.name
	return
}

func (t LoadInstance) String() (s string) {
	s = t.Name()
	return
}

func (t LoadInstance) Render() (item modules.Item) {
	item = RenderInstance(t)
	return
}

type LoadItem struct {
	Name   string `json:"name"`
	Text   string `json:"full_text"`
	Markup string `json:"markup"`
}

func (e LoadItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

type RenderContext struct {
	Load5, Load10, Load15 float32
}

func GetRenderContext() *load.AvgStat {
	if avg, err := load.Avg(); err == nil {
		return avg
	} else {
		log.Fatal(err)
	}
	return nil
}

func RenderInstance(i modules.ModuleInstance) (t modules.Item) {

	instance := i.(LoadInstance)
	var formatted bytes.Buffer

	renderContext := GetRenderContext()

	if instance.template == nil {
		log.Error("template is nil")
		return
	}

	if err := instance.template.Execute(&formatted, renderContext); err != nil {
		log.Fatal(err)
	}

	f := formatted.String()
	log.Debug(f)
	t = modules.Item(LoadItem{Name: instance.name, Text: f, Markup: "pango"})

	return
}

func color(value float64) string {
	numprocs := float64(runtime.NumCPU()) / 2
	switch {
	case value < 0.5*numprocs:
		return "grey"
	case value >= 0.5*numprocs && value < 1.5*numprocs:
		return "#D9FF00"
	case value >= 1.5*numprocs && value < 2*numprocs:
		return "yellow"
	case value >= 2*numprocs && value < 4*numprocs:
		return "orange"
	case value >= 4*numprocs:
		return "red"
	}
	return ""
}

func CreateInstance(name string, config map[string]interface{}) (m modules.ModuleInstance) {

	var format string

	if i, ok := config["format"]; ok {
		format = i.(string)
	} else {
		format = `<span color="{{ color .Load1 }}">{{ .Load1 | printf "%2.2f" }}</span> <span color="{{ color .Load5 }}">{{.Load5 | printf "%2.2f"}}</span> <span color="{{ color .Load15 }}">{{.Load15 | printf "%2.2f"}}</span>`
	}

	f := LoadInstance{
		name:   name,
		format: format,
	}

	if template, err := template.New(name).Funcs(template.FuncMap{
		"color": color,
	}).Parse(format); err == nil {
		f.template = template
	} else {
		log.Error("failed to create template: " + err.Error())
	}
	m = modules.ModuleInstance(f)

	return
}

var Module = modules.Module{
	Name:           "load",
	CreateInstance: CreateInstance,
}
