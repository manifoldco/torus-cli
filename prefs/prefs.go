package prefs

import (
	"os"
	"os/user"
	"path"
	"reflect"
	"strings"

	"github.com/manifoldco/torus-cli/errs"

	"github.com/go-ini/ini"
	"gopkg.in/oleiade/reflections.v1"
)

const (
	rcFilename        = ".torusrc"
	registryURI       = "https://registry.arigato.sh"
	manifestURI       = "https://get.torus.sh/manifest.json"
	gatekeeperAddress = "0.0.0.0:8200"
)

// Preferences represents the configuration as user has in their torusrc file
type Preferences struct {
	Core     Core     `ini:"core"`
	Defaults Defaults `ini:"defaults"`
}

// CountFields returns the number of defined fields on sub-field struct
func (prefs Preferences) CountFields(fieldName string) int {
	value, err := reflections.GetField(prefs, fieldName)
	count := 0
	if err != nil {
		return count
	}
	items, _ := reflections.Items(value)
	for i := range items {
		value, _ := reflections.GetField(value, i)
		if value != nil && value != "" {
			count++
		}
	}
	return count
}

// Core contains core option values
type Core struct {
	PublicKeyFile      string `ini:"public_key_file,omitempty"`
	CABundleFile       string `ini:"ca_bundle_file,omitempty"`
	RegistryURI        string `ini:"registry_uri,omitempty"`
	ManifestURI        string `ini:"manifest_uri,omitempty"`
	GatekeeperAddress  string `ini:"gatekeeper_address"`
	Context            bool   `ini:"context,omitempty"`
	AutoConfirm        bool   `ini:"auto_confirm,omitempty"`
	EnableProgress     bool   `ini:"progress"`
	EnableHints        bool   `ini:"hints"`
	Vim                bool   `ini:"vim,omitempty"`
	EnableCheckUpdates bool   `ini:"check_updates"`
}

// Defaults contains default values for use in command argument flags
type Defaults struct {
	Organization string `ini:"org,omitempty"`
	Project      string `ini:"project,omitempty"`
	Environment  string `ini:"environment,omitempty"`
	Service      string `ini:"service,omitempty"`
}

// SetValue for ini key on preferences struct
func (prefs Preferences) SetValue(key string, value string) (Preferences, error) {
	parts := strings.Split(key, ".")
	section := parts[0] // [Core|Default]
	key = parts[1]      // Rest of the property name

	// Identify category struct by ini tag name [Core|Default]
	values := reflect.ValueOf(&prefs).Elem()
	target := findElemByName(values, section)

	// Identify field to update by ini tag name
	values = reflect.ValueOf(&prefs).Elem().FieldByName(target)
	if values == reflect.Zero(reflect.TypeOf(values)).Interface() {
		return prefs, errs.NewExitError("error: unknown section `" + section + "`")
	}
	property := findElemByName(values, key)

	// Ensure the field is not zero
	field := values.Addr().Elem().FieldByName(property)
	if field == reflect.Zero(reflect.TypeOf(field)).Interface() {
		return prefs, errs.NewExitError("error: unknown property `" + key + "`")
	}

	switch field.Type() {
	case reflect.TypeOf(true):
		v := true
		if value != "true" && value != "1" {
			v = false
		}
		field.SetBool(v)
	default:
		field.SetString(value)
	}

	return prefs, nil
}

func findElemByName(values reflect.Value, iniField string) string {
	var fieldName string
	for i := 0; i < values.NumField(); i++ {
		tag := values.Type().Field(i).Tag
		names := strings.Split(tag.Get("ini"), ",")
		for _, a := range names {
			if a == iniField {
				fieldName = values.Type().Field(i).Name
			}
		}
	}
	return fieldName
}

// RcPath returns the torusrc filepath
func RcPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(u.HomeDir, rcFilename), nil
}

// NewPreferences returns a new instance of preferences struct
func NewPreferences() (*Preferences, error) {
	prefs := &Preferences{
		Core: Core{
			RegistryURI:        registryURI,
			ManifestURI:        manifestURI,
			GatekeeperAddress:  gatekeeperAddress,
			Context:            true,
			EnableHints:        true,
			EnableProgress:     true,
			EnableCheckUpdates: false,
		},
	}

	filePath, _ := RcPath()
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return prefs, nil
	}

	if err != nil {
		return prefs, err
	}

	rcPath, _ := RcPath()
	err = ini.MapTo(prefs, rcPath)
	return prefs, err
}
