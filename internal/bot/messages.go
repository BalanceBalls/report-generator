package bot

const helloMsg = `Доступные команды:
	/help - информация о боте
	/reg - зарегистрироваться в боте
	/unreg - удалить аккаунт
	/profile - информация об аккаунте
	/gen_day - сгенерировать репорт за день
`

const helpMsg = `
	Для того чтобы обновить или установить gitlab токен необходимо прислать 
	сообщение с префиксом 'token:gitlab_token' без пробелов.
	Пример:
	'token:glpat-wEp1SkMS_Yvr9kgDyt4A'

	Для того, чтобы обновить или установить gitlab идентификатор пользователя, необходимо прислать 
	сообщение с префиксом 'id:gitlab_id' без пробелов.
	Идентификатор можно найти в gitlab перейдя в Prefereneces -> Profile : User ID
	Пример:
	'id:123456789'

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
	tokenNotSetErrorMsg       = "Ошибка: gitlab токен не задан"
	gitlabIdNotSetErrorMsg    = "Ошибка: gitlab идентификатор пользователя не задан"
	gitlabIdBadInputErrorMsg  = "Ошибка: не удалось обработать полученный идентификатор пользователя gitlab"
	userDataUpdateErrorMsg    = "Ошибка при обновлении данных пользователя"
	userRegistrationErrorMsg  = "Ошибка при добавлении пользователя"
	reportGenerationFailedMsg = "Ошибка при создании отчета"
	fetchUserInfoFailedMsg    = "Ошибка при получении данных о пользователе"
	tokenHasBeenSavedMsg      = "Токен успешно сохранен"
	userHasBeenRemovedMsg     = "Аккаунт успешно удален"
	gitlabIdHasBeenSavedMsg   = "Gitlab идентификатор пользователя сохранен"
	userHasBeenRegisteredMsg  = "Пользователь успешно зарегестрирован. Необходимо обновить gitlab токен и идентификатор. Используйте /help для справки"
	timezoneHasBeenSavedMsg   = "Часовой пояс успешно сохранен"
	reportInProgressMsg       = "Отчет генерируется..."
	emptyReportMsg            = "Нет данных для отчета. Отсутствуют события в git"
	reportFileCaption         = "Отчет за сегодняшний день"
	tokenIsSetMsg             = "токен установлен"
	tokenIsNotSetMsg          = "токен не установлен"
)

const profileCmdTemplate = `
------Данные пользователя------
Часовой пояс: %d минут от GMT +0
Gitlab id: %d
Токен: %s
`
