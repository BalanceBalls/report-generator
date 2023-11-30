package bot

import "fmt"

type Config struct {
	DbName string `env:"DB_NAME" envDefault:"bot.sqlite"`

	PgPass string `env:"PG_PASS" envDefault:"admin"`
	PgUser string `env:"PG_USER" envDefault:"admin"`
	PgDb   string `env:"PG_DB" envDefault:"bot"`
	PgHost string `env:"PG_HOST" envDefault:"tg_bot_db"`

	ReportFileDir  string `env:"REPORT_FILE_DIR" envDefault:"./reports"`
	ReportTemplate string `env:"REPORT_TEMPLATE" envDefault:"html_report.tmpl"`
	GenerateFile   bool   `env:"GENERATE_FILE" envDefault:"false"`

	BotToken        string `env:"BOT_TOKEN,notEmpty"`
	CommandsTimeout int    `env:"COMMANDS_TIMEOUT" envDefault:"30"`

	GitHost     string `env:"GIT_HOST" envDefault:"localhost:4443"`
	GitBasePath string `env:"GIT_BASE_PATH" envDefault:"api/v4"`
}

func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", c.PgUser, c.PgPass, c.PgHost, c.PgDb)
}
