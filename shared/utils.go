package shared

import "gopkg.in/yaml.v2"

// MustYaml serializes i to a YAML string, or panics if it fails to do so.
func MustYaml(i interface{}) string {
	bytes, err := yaml.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
