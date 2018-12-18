package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
)

func GetHashedNameForFile(uID uint, filename string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(time.Now().String() + fmt.Sprintf("%v", uID) + filename))
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash[:16] + path.Ext(filename), nil
}

func SaveFile(file io.Reader, dir, filename string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	logger.Infow("saved file",
		"path", dir,
		"filename", filename)

	return nil
}
