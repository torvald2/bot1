package config

import "github.com/spf13/viper"

type Config struct {
	TOKEN           string     `json:"TOKEN"`
	ChannelID       int64      `json:"ChannelID"`
	Users           []string   `json:"Users"`
	Accounts        [][]string `json:"Accounts"`
	Subscribe       string     `json:"Subscribe"`
	OurChannels     string     `json:"OurChannels"`
	ContactUs       string     `json:"ContactUs"`
	TwitterLogin    string     `json:"TwitterLogin"`
	TwitterPassword string     `json:"TwitterPassword"`
	TwitterEmail    string     `json:"TwitterEmail"`
}

func New(fileName string) (*Config, error) {
	var conf *Config
	err := viperSetup(fileName)
	if err != nil {
		return &Config{}, err

	}

	err = viper.ReadInConfig()
	if err != nil {
		return &Config{}, err

	}
	err = viper.Unmarshal(&conf)
	if err != nil {
		return &Config{}, err
	}
	return conf, nil
}

func viperSetup(filename string) (error error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(filename)
	viper.SetConfigType("json")

	viper.AutomaticEnv()
	return nil
}
