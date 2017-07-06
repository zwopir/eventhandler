package cmd

import (
	"github.com/spf13/viper"
	"os"
	"reflect"
	"sort"
	"testing"
)

var (
	defaultConfig        = os.ExpandEnv("$PWD/testdata/config_example.yaml")
	defaultConfigSymlink = os.ExpandEnv("$HOME/config.yaml")
)

func getExpectedKeys() ([]string, error) {
	testConfig := viper.New()
	testConfig.SetConfigFile(defaultConfig)
	err := testConfig.ReadInConfig()
	if err != nil {
		return []string{}, err
	}
	ret := testConfig.AllKeys()
	sort.Strings(ret)
	return ret, nil
}

func copyToDefaultPath() error {
	return os.Symlink(defaultConfig, defaultConfigSymlink)
}

func rmDefaultPath() error {
	return os.Remove(defaultConfigSymlink)
}

func TestInitConfigDefaultPaths(t *testing.T) {
	expected, err := getExpectedKeys()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = copyToDefaultPath()
	if err != nil {
		t.Fatal(err.Error())
	}

	InitConfig()

	err = rmDefaultPath()
	if err != nil {
		t.Fatal(err.Error())
	}

	actual := viper.AllKeys()
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Config keys not loaded correctly. Expected:\n%+v\nActual:\n%+v", expected, actual)
	}
}
