package net

import (
	"bytes"
	"encoding/json"
	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	go_net "net"
	"text/template"
)

var log = logging.MustGetLogger("go3status.net")

type NetItem struct {
	Name   string `json:"name"`
	Text   string `json:"full_text"`
	Markup string `json:"markup"`
}

func (e NetItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

type NetInstance struct {
	name           string
	interface_name string
	template       *template.Template
	config         map[string]interface{}
}

func (t NetInstance) Name() (n string) {
	n = t.name
	return
}

func (t NetInstance) Config() (m map[string]interface{}) {
	m = t.config
	return
}

func (t NetInstance) String() (s string) {
	s = t.Name()
	return
}

type NetFormatData struct {
	Name           string
	Interface_name string
	Interface      *go_net.Interface
	Up             bool
	Addresses      []string
}

func (t NetInstance) Render() (i modules.Item) {

	if t.template == nil {
		log.Error("No template available.")
		return
	}
	_, linkLocalv6, _ := go_net.ParseCIDR("fe80::/10")
	_, privatev6, _ := go_net.ParseCIDR("fd00::/8")

	item := NetItem{Name: t.name, Markup: "pango"}

	interface_name := t.interface_name

	formatData := NetFormatData{Name: t.name, Interface_name: interface_name}

	if iface, err := go_net.InterfaceByName(interface_name); err == nil && iface != nil {
		formatData.Interface = iface
		formatData.Up = (iface.Flags & go_net.FlagUp) != 0
		if addrs, err := iface.Addrs(); err == nil && addrs != nil {
			for _, addr := range addrs {
				if ip, _, err := go_net.ParseCIDR(addr.String()); err == nil {
					if !linkLocalv6.Contains(ip) && !privatev6.Contains(ip) {
						formatData.Addresses = append(formatData.Addresses, addr.String())
					}
				}
			}
		}
	}

	var formatted bytes.Buffer
	t.template.Execute(&formatted, formatData)
	item.Text = formatted.String()

	i = modules.Item(item)
	return
}

func CreateInstance(name string, config map[string]interface{}) (moduleInstance modules.ModuleInstance) {
	i := NetInstance{
		name: name,
	}

	if v, ok := config["interface_name"]; ok {
		interface_name := v.(string)
		i.interface_name = interface_name

	} else {
		log.Error("Missing interface_name in " + name)
		moduleInstance = nil
		return
	}
	var format string
	if v, ok := config["format"]; ok {
		format = v.(string)
	} else {
		format = "{{.Interface_name}}: {{range $i, $v := .Addresses}}{{if $i}}, {{end}}{{$v}}{{end}}"
	}
	if template, err := template.New(i.name).Parse(format); err == nil {
		i.template = template
	} else {
		log.Error("Failed to create template: " + err.Error())
	}

	moduleInstance = i

	return
}

var Module = modules.Module{
	Name:           "net",
	CreateInstance: CreateInstance,
}
