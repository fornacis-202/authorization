package conf

import (
	"fmt"
	"log"

	"ImageProccessor/pkg/imagga"
	"ImageProccessor/pkg/mailgun"
	"ImageProccessor/pkg/mqtt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

// Config
// struct type of app configs.
type Config struct {
	Imagga  imagga.Config  `koanf:"imagga"`
	Mailgun mailgun.Config `koanf:"mailgun"`
	MQTT    mqtt.Config    `koanf:"mqtt"`
	EncKey  string         `koanf:"encKey"`
	S3      s3Config       `koanf:"s3"`
	Mysql   mysqlConfig    `koanf:"mysql"`
}

// Load
// loading app configs.
func Load() Config {
	var instance Config

	k := koanf.New(".")

	// load default
	if err := k.Load(structs.Provider(Default(), "koanf"), nil); err != nil {
		_ = fmt.Errorf("error loading deafult: %v\n", err)
	}

	// load configs file
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		_ = fmt.Errorf("error loading config.yaml file: %v\n", err)
	}

	// unmarshalling
	if err := k.Unmarshal("", &instance); err != nil {
		log.Fatalf("error unmarshalling config: %v\n", err)
	}

	return instance
}

func Default() Config {
	return Config{
		Imagga: imagga.Config{
			ApiKey:    "",
			ApiSecret: "",
		},
		Mailgun: mailgun.Config{
			Domain: "",
			APIKEY: "",
			Sender: "",
		},
		MQTT: mqtt.Config{
			Queue: "",
			URI:   "",
		},

		EncKey: "",
		Mysql: mysqlConfig{
			DBName: "",
			User:   "",
			Passwd: "",
			Addr:   "",
		},
		S3: s3Config{
			AccessKey: "",
			SecretKey: "",
			Region:    "",
			Bucket:    "",
			Endpoint:  "",
		},
	}
}

type s3Config struct {
	AccessKey string `koanf:"accessKey"`
	SecretKey string `koanf:"secretKey"`
	Region    string `koanf:"region"`
	Bucket    string `koanf:"bucket"`
	Endpoint  string `koanf:"endpoint"`
}

type mysqlConfig struct {
	DBName string `koanf:"dbName"`
	User   string `koanf:"user"`
	Passwd string `koanf:"passwd"`
	Addr   string `koanf:"addr"`
}
