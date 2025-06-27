package util

import (
	"github.com/spf13/viper"
)

//* this config file stores all config of application and value
//* are read by viper and from a config/env file

//& what actually it does is to unmarshall into embedded struts
//& uses the mapstrcutre under the hood so we have to use it

type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	//! it autmoaticaly over writes the values it read from the config file
	//! with the values in the ENV if they exist
	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
