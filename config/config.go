package config

import "github.com/spf13/viper"

type Config struct {
	App struct {
		ServerHost string `mapstructure:"server_host"`
		GRPCPort   int    `mapstructure:"grpc_port"`
	} `mapstructure:"app"`

	Services struct {
		UserPort int `mapstructure:"user_port"`
	} `mapstructure:"services"`

	Database struct {
		DBHost           string `mapstructure:"pg_host"`
		DBName           string `mapstructure:"pg_database"`
		DBUser           string `mapstructure:"pg_user"`
		DBPassword       string `mapstructure:"pg_password"`
		DBSSLMode        string `mapstructure:"pg_ssl_mode"`
		DBChannelBinding string `mapstructure:"pg_channel_binding"`
	} `mapstructure:"database"`

	MessageQueue struct {
		RHost     string `mapstructure:"rb_host"`
		RUser     string `mapstructure:"rb_user"`
		RPassword string `mapstructure:"rb_password"`
	} `mapstructure:"message_queue"`

	ImageKit struct {
		PublicKey   string `mapstructure:"public_key"`
		PrivateKey  string `mapstructure:"private_key"`
		URLEndpoint string `mapstructure:"url_endpoint"`
		Folder      string `mapstructure:"folder"`
	} `mapstructure:"imagekit"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config/config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
