package dependencies

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config interface {
	Get(string) (any, error)
	GetString(string) (*string, error)
	GetInteger(string) (*int, error)
}

type ViperConfig struct {
	config *viper.Viper
}

func NewViperConfig() (Config, error) {
	vp := viper.New()
	vp.SetConfigName("config")
	vp.AddConfigPath(".")
	vp.AutomaticEnv()
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}
	return &ViperConfig{config: vp}, nil
}

func (cf *ViperConfig) Get(nestedToken string) (any, error) {
	res := cf.config.Get(nestedToken)
	if res == nil {
		return nil, fmt.Errorf("%s not provided!", nestedToken)
	}
	return res, nil
}

func (cf *ViperConfig) GetString(token string) (*string, error) {
	res, err := cf.Get(token)
	if err != nil {
		return nil, err
	}

	if stringValue, ok := res.(string); ok {
		return &stringValue, nil
	}
	return nil, fmt.Errorf("Cannot read the string value")
}

func (cf *ViperConfig) GetInteger(token string) (*int, error) {
	res, err := cf.Get(token)
	if err != nil {
		return nil, err
	}
	if intValue, ok := res.(int); ok {
		return &intValue, nil
	}
	return nil, fmt.Errorf("Cannot read the integer value")
}
