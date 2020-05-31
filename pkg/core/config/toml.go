package config

import "github.com/BurntSushi/toml"

// DecodeFile decode toml file
func DecodeFile(fpath string, v interface{}) (toml.MetaData, error) {
	return toml.DecodeFile(fpath, v)
}
