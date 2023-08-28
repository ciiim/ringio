package conf

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfig() *viper.Viper {
	v := viper.New()
	v.SetConfigName("conf")
	v.SetConfigType("toml")
	v.AddConfigPath("./conf")
	err := v.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Make a default config
		log.Println("Config file not found, will create one.")

		v.Set("server.server_name", "")
		v.Set("server.api_server_port", "8080")

		v.Set("conf.use", "")
		v.Set("debug", false)

		v.Set("database.mysql.datasource", "")

		v.WriteConfigAs("conf/conf.toml")
		log.Fatalln("Please edit conf/conf.toml and start the program again.")
	} else if err != nil {
		log.Fatalln("Reading config file error:", err)
	}
	setConf(v)
	return v
}

func setConf(v *viper.Viper) {
	mode := v.GetString("conf.use")
	if mode == "" {
		log.Println("Using Default Configuration.")
		return
	}
	switch mode {
	case "dev":
		log.Println("Using Development Configuration.")
		v.SetConfigName("dev")
		err := v.ReadInConfig()
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("%s.toml not found.", mode)
		}
	case "pro":
		log.Println("Using Production Configuration.")
		v.SetConfigName("pro")
		err := v.ReadInConfig()
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("%s.toml not found.", mode)
		}
	default:
		log.Println("Using Custom Configuration.")
		v.SetConfigName(mode)
		err := v.ReadInConfig()
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("%s.toml not found.", mode)
		}
	}
}

func GetNodes(v *viper.Viper) []map[string]string {
	nodes, ok := v.Get("cluster.nodes").([]any)
	if !ok {
		return nil
	}
	var nodeList []map[string]string
	for _, node := range nodes {
		nodeMap, ok := node.(map[string]any)
		if !ok {
			continue
		}
		nodeList = append(nodeList, map[string]string{
			"name": nodeMap["name"].(string),
			"ip":   nodeMap["ip"].(string),
		})
	}
	return nodeList
}
