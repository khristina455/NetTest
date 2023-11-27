package db

import (
	"github.com/spf13/viper"
)

func GetConnectionString() (string, error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}
	return viper.GetString("db.connection_string"), nil
}
