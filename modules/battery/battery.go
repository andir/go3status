package battery

import (
	"bytes"
	"encoding/json"
	"github.com/andir/go3status/modules"
	"github.com/op/go-logging"
	"io/ioutil"
	//"reflect"
	"strconv"
	"strings"
	"text/template"
)

var log = logging.MustGetLogger("go3status.battery")

type BatteryItem struct {
	Name   string `json:"name"`
	Text   string `json:"full_text"`
	Markup string `json:"markup"`
}

func (e BatteryItem) Marshal() (bytes []byte) {
	var err error
	if bytes, err = json.Marshal(e); err != nil {
		log.Error(err.Error())
	}
	return
}

type BatteryInstance struct {
	name        string
	device_path string
	template    *template.Template
}

func (i BatteryInstance) Name() string {
	return i.name
}

func (i BatteryInstance) String() (s string) {
	s = i.name
	s += " "
	s += i.device_path
	return
}

type BatteryInfo struct {
	Name               string  `json:"Name"`
	Status             string  `json:"Status"`
	Present            string  `json:"Present"`
	Technology         string  `json:"Technology"`
	Cycle_count        int     `json:"Cycle_Count"`
	Voltage_min_design int     `json:"Voltage_Min_Design"`
	Voltage_now        int     `json:"Voltage_Now"`
	Power_now          int     `json:"Power_Now"`
	Energy_full_design int     `json:"Energy_full_design"`
	Energy_full        int     `json:"Energy_full"`
	Energy_now         int     `json:"Energy_now"`
	Capacity           int     `json:"Capacity"`
	Capacity_level     string  `json:"Capacity_level"`
	Model_name         string  `json:"Module_name"`
	Manufacturer       string  `json:"Manufacturer"`
	Serial_number      string  `json:"Serial_number"`
	Percentage         float64 `json:"Percentage"`
}

func NewBatteryInfo(fileName string) *BatteryInfo {
	info := &BatteryInfo{}

	if v, err := ioutil.ReadFile(fileName); err != nil {
		log.Error(err.Error())
		return nil
	} else {
		s := string(v)
		s = strings.TrimSpace(s)
		for _, line := range strings.Split(s, "\n") {
			line = strings.TrimSpace(line)
			tokens := strings.SplitN(line, "=", 2)
			if len(tokens) != 2 {
				log.Error("Failed to parse line: " + line)
				continue
			}
			key := tokens[0]
			val := tokens[1]
			switch key {
			case "POWER_SUPPLY_NAME":
				info.Name = val
			case "POWER_SUPPLY_STATUS":
				info.Status = val
			case "POWER_SUPPLY_PRESENT":
				info.Present = val
			case "POWER_SUPPLY_TECHNOLOGY":
				info.Technology = val
			case "POWER_SUPPLY_CYCLE_COUNT":
				info.Cycle_count, _ = strconv.Atoi(val)
			case "POWER_SUPPLY_VOLTAGE_MIN_DESIGN":
				if v, err := strconv.Atoi(val); err == nil {
					info.Voltage_min_design = v
				}
			case "POWER_SUPPLY_VOLTAGE_NOW":
				if v, err := strconv.Atoi(val); err == nil {
					info.Voltage_now = v
				}
			case "POWER_SUPPLY_POWER_NOW":
				if v, err := strconv.Atoi(val); err == nil {
					info.Power_now = v
				}
			case "POWER_SUPPLY_ENERGY_FULL_DESIGN":
				if v, err := strconv.Atoi(val); err == nil {
					info.Energy_full_design = v
				}
			case "POWER_SUPPLY_ENERGY_FULL":
				if v, err := strconv.Atoi(val); err == nil {
					info.Energy_full = v
				}
			case "POWER_SUPPLY_ENERGY_NOW":
				if v, err := strconv.Atoi(val); err == nil {
					info.Energy_now = v
				}
			case "POWER_SUPPLY_CAPACITY":
				if v, err := strconv.Atoi(val); err == nil {
					info.Capacity = v
				}
			case "POWER_SUPPLY_CAPACITY_LEVEL":
				info.Capacity_level = val
			case "POWER_SUPPLY_MODEL_NAME":
				info.Model_name = val
			case "POWER_SUPPLY_MANUFACTURER":
				info.Manufacturer = val
			case "POWER_SUPPLY_SERIAL_NUMBER":
				info.Serial_number = val
			}

			charge := float64(info.Energy_now) / float64(info.Energy_full)
			charge *= 100
			info.Percentage = charge
		}
	}
	return info
}

func (i BatteryInstance) Render() (item modules.Item) {
	it := BatteryItem{
		Name: i.name,
	}

	if i.template == nil {
		log.Error("No template available.")
		item = nil
		return
	}

	info := NewBatteryInfo(i.device_path)

	if b, err := json.Marshal(info); err != nil {
		log.Error("Failed to marshal info.")
	} else {
		log.Debug(string(b))
	}

	buffer := bytes.Buffer{}
	if err := i.template.Execute(&buffer, info); err != nil {
		log.Error(err.Error())
		item = nil
		return
	} else {
		it.Text = buffer.String()
	}

	item = it
	return
}

func CreateInstance(name string, config map[string]interface{}) (instance modules.ModuleInstance) {
	batteryInstance := BatteryInstance{
		name: name,
	}

	if v, ok := config["device_path"]; ok {
		batteryInstance.device_path = v.(string)
	} else {
		batteryInstance.device_path = "/sys/class/power_supply/BAT0/uevent"
	}

	var format string

	if v, ok := config["format"]; ok {
		format = v.(string)
	} else {
		format = `{{.Name}}: {{printf "%.1f" .Percentage}} % {{ if Equal .Status "Charging" }}⚇{{ end }}`
	}

	if tmpl, err := template.New(name).Funcs(template.FuncMap{
		"Equal": strings.EqualFold,
	}).Parse(format); err == nil {
		batteryInstance.template = tmpl
	} else {
		log.Error(err.Error())
	}

	instance = batteryInstance
	return
}

var Module = modules.Module{
	Name:           "battery",
	CreateInstance: CreateInstance,
}
