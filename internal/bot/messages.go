package bot

const helloMsg = `Доступные команды:
	/help - информация о боте
	/reg - зарегистрироваться в боте
	/unreg - удалить аккаунт
	/gen_day - сгенерировать репорт за день
`

const helpMsg = `
	Для того чтобы обновить или установить gitlab токен необходимо прислать 
	сообщение с префиксом 'token:gitlab_token' без пробелов.
	Пример:
	'token:glpat-wEp1SkMS_Yvr9kgDyt4A'

	Для того чтобы обновить или установить часовой пояс необходимо прислать 
	сообщение с префиксом 'offset:min_from_utc' без пробелов.
	Вместо 'min_from_utc' необходимо разницу от UTC в минутах.
	Примеры:
	UTC +5 (ЕКБ) = 'offset:300'
	UTC -5 (Нью-Йорк) = 'offset:-300'
`

// replies
const (
	userNotRegisteredMsg      = "Ошибка: пользователь не зарегестрирован. Для регистрации воспользуйтесь командой /reg"
	userAlreadyRegisteredMsg  = "Ошибка: пользователь уже зарегистрирован"
	userDataUpdateErrorMsg    = "Ошибка при обновлении данных пользователя"
	userRegistrationErrorMsg  = "Ошибка при добавлении пользователя"
	reportGenerationFailedMsg = "Ошибка при создании отчета"
	tokenHasBeenSavedMsg      = "Токен успешно сохранен"
	userHasBeenRemovedMsg     = "Аккаунт успешно удален"
	userHasBeenRegisteredMsg  = "Пользователь успешно зарегестрирован. Необходимо обновить gitlab токен. Используйте /help для справки"
	timezoneHasBeenSavedMsg   = "Часовой пояс успешно сохранен"
	reportInProgressMsg       = "Отчет генерируется..."
	emptyReportMsg            = "Нет данных для отчета. Отсутствуют события в git"
)
