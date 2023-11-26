package bot

type Config struct {
	DbName          string `env:"DB_NAME" envDefault:"bot.sqlite"`
	ReportFileDir   string `env:"REPORT_FILE_DIR" envDefault:"./reports"`
	ReportTemplate  string `env:"REPORT_TEMPLATE" envDefault:"html_report.tmpl"`
	GenerateFile    bool   `env:"GENERATE_FILE" envDefault:"false"`
	BotToken        string `env:"BOT_TOKEN,notEmpty"`
	CommandsTimeout int    `env:"COMMANDS_TIMEOUT" envDefault:"30"`
	GitHost         string `env:"GIT_HOST" envDefault:"localhost:4443"`
	GitBasePath     string `env:"GIT_BASE_PATH" envDefault:"api/v4"`
}
