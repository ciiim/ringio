package conf

import "github.com/spf13/viper"

func InitConfig() *viper.Viper {
	v := viper.New()
	v.SetConfigName("dev")
	v.SetConfigType("toml")
	v.AddConfigPath("./conf")
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return v
}
