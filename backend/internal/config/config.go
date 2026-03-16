package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv          string `mapstructure:"APP_ENV"`
	Port            string `mapstructure:"PORT"`
	RedisAddr       string `mapstructure:"REDIS_ADDR"`
	RedisPassword   string `mapstructure:"REDIS_PASSWORD"`
	JWTSecret       string `mapstructure:"JWT_SECRET"`
	CleanupTime     string `mapstructure:"CLEANUP_TIME"`
	CleanupTimezone string `mapstructure:"CLEANUP_TIMEZONE"`
	AdminUsername   string `mapstructure:"ADMIN_USERNAME"`
	AdminPassword   string `mapstructure:"ADMIN_PASSWORD"`

	BilibiliImportEnabled        bool    `mapstructure:"BILIBILI_IMPORT_ENABLED"`
	BilibiliImportKeyword        string  `mapstructure:"BILIBILI_IMPORT_KEYWORD"`
	BilibiliImportLimit          int     `mapstructure:"BILIBILI_IMPORT_LIMIT"`
	BilibiliImportMinPrice       float64 `mapstructure:"BILIBILI_IMPORT_MIN_PRICE"`
	BilibiliImportSearchPages    int     `mapstructure:"BILIBILI_IMPORT_SEARCH_PAGES"`
	BilibiliImportSearchPageSize int     `mapstructure:"BILIBILI_IMPORT_SEARCH_PAGE_SIZE"`
	BilibiliImportCommentPages   int     `mapstructure:"BILIBILI_IMPORT_COMMENT_PAGES"`
	BilibiliImportTimeoutSeconds int     `mapstructure:"BILIBILI_IMPORT_TIMEOUT_SECONDS"`
	BilibiliCookie               string  `mapstructure:"BILIBILI_COOKIE"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_ENV", "dev")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_SECRET", "lingbao-secret-key-change-me")
	viper.SetDefault("CLEANUP_TIME", "00:00")
	viper.SetDefault("CLEANUP_TIMEZONE", "Local")
	viper.SetDefault("ADMIN_USERNAME", "")
	viper.SetDefault("ADMIN_PASSWORD", "")

	viper.SetDefault("BILIBILI_IMPORT_ENABLED", true)
	viper.SetDefault("BILIBILI_IMPORT_KEYWORD", "小马糕")
	viper.SetDefault("BILIBILI_IMPORT_LIMIT", 30)
	viper.SetDefault("BILIBILI_IMPORT_MIN_PRICE", 900)
	viper.SetDefault("BILIBILI_IMPORT_SEARCH_PAGES", 1)
	viper.SetDefault("BILIBILI_IMPORT_SEARCH_PAGE_SIZE", 20)
	viper.SetDefault("BILIBILI_IMPORT_COMMENT_PAGES", 1)
	viper.SetDefault("BILIBILI_IMPORT_TIMEOUT_SECONDS", 60)
	viper.SetDefault("BILIBILI_COOKIE", "")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Clean up port if it has a colon
	if strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = cfg.Port[1:]
	}

	return &cfg, nil
}
