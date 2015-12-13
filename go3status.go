package main

import (
	"encoding/json"
	"fmt"
	modules "github.com/andir/go3status/modules"
	go3_battery "github.com/andir/go3status/modules/battery"
	go3_idlerpg "github.com/andir/go3status/modules/idlerpg"
	go3_mpd "github.com/andir/go3status/modules/mpd"
	go3_net "github.com/andir/go3status/modules/net"
	go3_time "github.com/andir/go3status/modules/time"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var log = logging.MustGetLogger("go3status")

type Element struct {
	Color                 string `json:"color"`
	Name                  string `json:"name"`
	Full_text             string `json:"full_text"`
	Markup                string `json:"markup"`
	Seperator_block_width int    `json:"seperator_block_width"`
}

func setupLogging() {
	var format = logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
	)
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter)
}

func parseModuleConfig(name string, moduleConfig map[string]interface{}, mods map[string]modules.Module) (instance modules.ModuleInstance) {
	var mod modules.Module
	var ok bool
	var val interface{}
	val, ok = moduleConfig["module"]
	if !ok {
		//err = error()
		return
	}
	modname := val.(string)
	log.Debug("module:" + string(modname))

	if mod, ok = mods[modname]; !ok {
		log.Error("Couldn't find module: " + modname)
		return
	}
	instance = mod.CreateInstance(name, moduleConfig)

	return
}

func parseConfig(config string, mods map[string]modules.Module) (instances []modules.ModuleInstance) {
	var m map[string]interface{}

	if err := json.Unmarshal([]byte(config), &m); err == nil {
		for key, value := range m {
			log.Debug(key)
			switch value.(type) {
			case map[string]interface{}:
				instance := parseModuleConfig(key, value.(map[string]interface{}), mods)
				if instance != nil {
					log.Debug("Created instance:", instance.String())
					instances = append(instances, instance)
				} else {
					log.Debug("Failed to parse config for ", key)
				}
			default:
				log.Error("Failed to parse " + key)
			}
		}
	} else {
		log.Error(err.Error())
	}
	return
}

type CacheEntry struct {
	ts   time.Time
	item modules.Item
}

var cache = make(map[string]CacheEntry)

func render(instances []modules.ModuleInstance) {
	s := []string{}
	for _, instance := range instances {
		name := instance.Name()
		log.Info(name)
		var item modules.Item
		if v, ok := cache[name]; ok {
			if int(time.Since(v.ts).Seconds()) >= instance.RefreshInterval() {
				item = instance.Render()
				cache[name] = CacheEntry{ts: time.Now(), item: item}
			} else {
				item = v.item
			}
		} else {
			item = instance.Render()
			cache[name] = CacheEntry{ts: time.Now(), item: item}
		}
		if item != nil {
			s = append(s, string(item.Marshal()))
		} else {
			log.Error(instance.Name() + " did not return a valid item")
		}
	}
	fmt.Println("[\n" + strings.Join(s, ",\n") + "],\n")
}

type Run struct {
	val bool
}

func mainLoop(interval int64, instances []modules.ModuleInstance, run *Run) {

	/*
		*
		*
		{"stop_signal": 20, "click_events": true, "version": 1, "cont_signal": 18}
		[
		[],
		[{"color": "#AAAAAA", "separator_block_width": 0, "name": "traffic-wl0_rx", "markup": "pango", "full_text": "_",
		 "separator": false}, {"color": "#AAAAAA", "name": "traffic-wl0_tx", "markup": "pango", "full_text": "_"},
		 {"color": "#FFFFFF", "name": "wireless_default", "full_text": "darmstadt.freifunk.net"},
		 {"color": "#AAAAAA", "name": "datetime_default", "full_text": "01:33:59"}
		],

		*
	*/
	var preamble = `{"stop_signal": 20, "click_events": true, "version": 1, "cont_signal": 18}`
	fmt.Println(preamble)

	fmt.Println("[")
	fmt.Println("[],")

	render(instances)
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		if run.val {
			render(instances)
		}
	}
}

func main() {
	var mods = map[string]modules.Module{}
	setupLogging()

	log.Info("Yay! Lets rock!")

	mods["time"] = go3_time.Module
	mods["net"] = go3_net.Module
	mods["mpd"] = go3_mpd.Module
	mods["battery"] = go3_battery.Module
	mods["idlerpg"] = go3_idlerpg.Module
	var config string

	if len(os.Args) > 1 {
		if text, err := ioutil.ReadFile(os.Args[1]); err == nil {
			config = string(text)
		} else {
			log.Error(err.Error())
		}
	} else {
		config = `
{
	"idlerpg-andi": {
		"module": "idlerpg",
		"player": "andi-"
	},
	"idlerpg-hexa": {
		"module": "idlerpg",
		"player": "hexa"
	},
	"local_mpd": {
		"module": "mpd",
		"format": "MPD: [{{ .State }}] {{ .Artist }} - {{ .Title }}"
	},
	"wireless_network": {
		"module": "net",
		"interface_name": "wlp4s0",
		"format": "<span color=\"{{ if .Up }}green{{ else }}red{{end}}\">{{.Interface_name}}</span>: {{range $i, $v := .Addresses}}{{if $i}}, {{end}}{{$v}}{{end}}"
	},
	"default_time": {
		"module": "time"
	},
	"default_battery": {
		"module": "battery"
	}
}`
	}
	instances := parseConfig(config, mods)
	if len(instances) == 0 {
		log.Error("No instances configured, exiting.")

	} else {

		var run = &Run{true}

		sigs := make(chan os.Signal)

		signal.Notify(sigs, syscall.SIGSTOP, syscall.SIGCONT)

		go func(m *Run) {
			for {
				sig := <-sigs
				if sig == syscall.SIGSTOP {
					m.val = false
				} else if sig == syscall.SIGCONT {
					m.val = true
				}
			}
		}(run)

		mainLoop(1, instances, run)
	}
}
