package filesystem

import (
	"net/http"
)

type StaticManager struct {
	fs http.Handler
}

// @Summary Отдать файл
// @Description Отдать файл с диска
// @ID get-static
// @Param PathToFile path string true "Путь к файлу"
// @Success 200 "Файл найден"
// @Failure 301 "Редирект, если имя папки не заканчивается на /"
// @Failure 403 "Нет прав (сервер)"
// @Failure 404 "Файл не найден"
// @Failure 500 "Внутренняя ошибка"
// @Router /static/{path/to/file} [GET]
func (stm *StaticManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stm.fs.ServeHTTP(w, r)
}

func NewStaticManager(path, directory string) *StaticManager {
	return &StaticManager{
		fs: http.StripPrefix(path, http.FileServer(http.Dir(directory))),
	}
}
