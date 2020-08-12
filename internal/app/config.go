package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// LoadConfig loads the global application configuration from a config
// file or if available from environment variables.
func LoadConfig() *Config {
	setDefaults()
	loadConfig()

	return &Config{
		App: appConfig{
			Locale:      viper.GetString("app.locale"),
			Environment: viper.GetString("app.env"),
		},
		Server: serverConfig{
			Host:       viper.GetString("server.host"),
			Port:       viper.GetString("server.port"),
			RootDir:    viper.GetString("server.root_dir"),
			StaticDir:  viper.GetString("server.static_dir"),
			LocalesDir: viper.GetString("server.locale_dir"),
			StorageDir: viper.GetString("server.storage_dir"),
		},
		Database: dbConfig{
			Host:       viper.GetString("database.host"),
			Name:       viper.GetString("database.name"),
			Port:       viper.GetString("database.port"),
			Username:   viper.GetString("database.username"),
			Password:   viper.GetString("database.password"),
			Migrations: viper.GetString("database.migrations"),
			Backup:     viper.GetBool("database.backup"),
			BackupTime: viper.GetString("database.backup_time"),
		},
		Log: logConfig{
			Level: viper.GetString("log.level"),
			Dir:   viper.GetString("log.dir"),
		},
	}
}

// Config represents the global application configuration.
type Config struct {
	App      appConfig
	Server   serverConfig
	Gets     getsConfig
	Database dbConfig
	Log      logConfig
}

type appConfig struct {
	Environment string
	Locale      string
}

type getsConfig struct {
	IP       string
	Username string
	Password string
	ReadOnly bool
}

type serverConfig struct {
	Host       string
	Port       string
	RootDir    string
	StaticDir  string
	LocalesDir string
	StorageDir string
}

func (s *serverConfig) URL() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

type dbConfig struct {
	Host       string
	Name       string
	Port       string
	Username   string
	Password   string
	Path       string
	Migrations string
	Backup     bool
	BackupTime string
}

func (config *dbConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@(%s:%s)/%s?multiStatements=true&parseTime=true&loc=UTC&collation=utf8mb4_general_ci",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)
}

type logConfig struct {
	Level string
	Dir   string
}

func setDefaults() {
	env := "develop"
	if envEnv := os.Getenv("EXAMPLE_ENV"); envEnv != "" {
		env = envEnv
	}

	viper.SetDefault("app.locale", "de")
	viper.SetDefault("app.env", env)

	viper.SetDefault("server.host", "")
	viper.SetDefault("server.port", "80")
	viper.SetDefault("server.static_dir", "/app/static")
	viper.SetDefault("server.root_dir", "/app")
	viper.SetDefault("server.locale_dir", "/app/locales")
	viper.SetDefault("server.storage_dir", "/go-webapp-example/data/storage")

	viper.SetDefault("database.host", "db")
	viper.SetDefault("database.name", "gowebapp")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.username", "gowebapp")
	viper.SetDefault("database.password", "gowebapp")
	viper.SetDefault("database.migrations", "/app/migrations")
	viper.SetDefault("database.backup", true)
	viper.SetDefault("database.backup_time", "03:00")

	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.dir", "/go-webapp-example/log")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("/go-webapp-example/data")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %s", err))
	}

	viper.WatchConfig()
	viper.SetEnvPrefix("CS")
}
