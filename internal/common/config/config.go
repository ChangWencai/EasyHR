package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OSS      OSSConfig      `mapstructure:"oss"`
	SMS      SMSConfig      `mapstructure:"sms"`
	Crypto   CryptoConfig   `mapstructure:"crypto"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
}

type SMSConfig struct {
	Provider        string `mapstructure:"provider"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"`
}

type CryptoConfig struct {
	AESKey string `mapstructure:"aes_key"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnv(v, "database.password", "DATABASE_PASSWORD")
	bindEnv(v, "jwt.secret", "JWT_SECRET")
	bindEnv(v, "crypto.aes_key", "AES_KEY")
	bindEnv(v, "oss.access_key_id", "OSS_ACCESS_KEY_ID")
	bindEnv(v, "oss.access_key_secret", "OSS_ACCESS_KEY_SECRET")
	bindEnv(v, "sms.access_key_id", "SMS_ACCESS_KEY_ID")
	bindEnv(v, "sms.access_key_secret", "SMS_ACCESS_KEY_SECRET")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func bindEnv(v *viper.Viper, key, envVar string) {
	_ = v.BindEnv(key, envVar)
}
