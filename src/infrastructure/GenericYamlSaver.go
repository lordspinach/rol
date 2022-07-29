package infrastructure

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"rol/app/errors"
)

func createYamlFile(fileName string, obj interface{}) error {
	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, yamlData, 0664)
	if err != nil {
		return err
	}
	return nil
}

//SaveYamlFile saving struct to a yaml file along the given path
//
//Return:
//	error - if an error occurs, otherwise nil
func SaveYamlFile[StructType interface{}](obj StructType, filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		createErr := createYamlFile(filePath, new(StructType))
		if createErr != nil {
			return errors.Internal.Wrap(createErr, "error when creating a yaml file")
		}
	}
	err := createYamlFile(filePath, obj)
	if err != nil {
		return err
	}

	return nil
}

//ReadYamlFile reads the yaml file at the given path into the specified struct
//
//Return:
//	StructType - specified struct with yaml data
//	error - if an error occurs, otherwise nil
func ReadYamlFile[StructType interface{}](filePath string) (StructType, error) {
	config := new(StructType)
	if _, err := os.Stat(filePath); err != nil {
		createErr := createYamlFile(filePath, config)
		if createErr != nil {
			return *config, fmt.Errorf("%s: %s", err, createErr)
		}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return *config, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return *config, err
	}

	return *config, nil
}
