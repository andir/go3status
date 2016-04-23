package load

import (
	"encoding/json"
	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	mem "github.com/shirou/gopsutil/mem"
	humanize "github.com/dustin/go-humanize"
	"text/template"
	"bytes"
)

var log = logging.MustGetLogger("go3status.memory")

type MemoryInstance struct {
	name   string
	format string
	template *template.Template
}

func (t MemoryInstance) RefreshInterval() int {
	return 5
}

func (t MemoryInstance) Name() (n string) {
	n = t.name
	return
}

func (t MemoryInstance) String() (s string) {
	s = t.Name()
	return
}

func (t MemoryInstance) Render() (item modules.Item) {
	item = RenderInstance(t)
	return
}

type LoadItem struct {
	Name string `json:"name"`
	Text string `json:"full_text"`
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


func GetRenderContext() *mem.VirtualMemoryStat {
	if avg, err := mem.VirtualMemory(); err == nil {
		return avg
	} else {
		log.Fatal(err)
	}
	return nil
}

func RenderInstance(i modules.ModuleInstance) (t modules.Item) {

	instance := i.(MemoryInstance)
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
	t = modules.Item(LoadItem{Name: instance.name, Text: f, Markup:"pango"})

	return
}

func convert(value uint64) string {

	return humanize.IBytes(value)
}

func CreateInstance(name string, config map[string]interface{}) (m modules.ModuleInstance) {

	var format string

	if i, ok := config["format"]; ok {
		format = i.(string)
	} else {
		format = `Memory: {{ printf "%3.2f %%" .UsedPercent }} ({{ convert .Used }} / {{ convert .Total }})`
	}

	f := MemoryInstance{
		name:   name,
		format: format,
	}

	funcMap := template.FuncMap{
		"convert": convert,
	}

	if template, err := template.New(name).Funcs(funcMap).Parse(format); err == nil {
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
