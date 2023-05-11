package usercookies

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-url-shortener/internal/logger"
	"net/http"
)

// название cookie в которой храняться данные пользователя
var NameCookiesUserData = "USER_DATA"

// тип который описывает данные пользователя, хранящиеся в куках
type UserDataCookies struct {
	ListFullURL []string
}

// Получаем из кук данне пользователя
func GetCookiesUserData(req *http.Request) (userData UserDataCookies, err error) {

	cookiesLinksUsers := ""
	if valueCookie, err := req.Cookie(NameCookiesUserData); err == nil {
		cookiesLinksUsers = valueCookie.Value
	}

	if cookiesLinksUsers != "" {
		bytesCookiesLinksUsers, err := base64.StdEncoding.DecodeString(cookiesLinksUsers)
		err = json.Unmarshal(bytesCookiesLinksUsers, &userData)
		if err != nil {
			err = fmt.Errorf("ошибка десериализации куков "+NameCookiesUserData+": %w", err)
			strError := err.Error()
			logger.GetLogger().Debugf("%s", strError)
		}
	}
	return
}

// Сохраняем в куках данные пользователя
func SetCookiesUserData(userData UserDataCookies, res http.ResponseWriter) {
	bytesCookiesLinksUsers, _ := json.Marshal(&userData)
	encodenCookiesLinksUsers := base64.StdEncoding.EncodeToString(bytesCookiesLinksUsers)
	cookie := &http.Cookie{
		Name:   NameCookiesUserData,
		Value:  encodenCookiesLinksUsers,
		Path:   "/",
		MaxAge: 30000,
		Secure: false,
	}
	http.SetCookie(res, cookie)
}

// Запоминаем ссылку, которую запрашивал пользователь
func AddFullURLToUser(fullURL string, res http.ResponseWriter, req *http.Request) (err error) {

	userData, err := GetCookiesUserData(req)
	logger.GetLogger().Debugf("Ссылки текущего пользователя до добавления %+v", userData)
	if err != nil {
		return err
	}

	isNewLink := true
	for _, value := range userData.ListFullURL {
		if value == fullURL {
			isNewLink = false
		}
	}
	if isNewLink {
		userData.ListFullURL = append(userData.ListFullURL, fullURL)
	}

	SetCookiesUserData(userData, res)
	return
}
