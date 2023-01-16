package log

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/ghnexpress/traefik-cache/model"
)

type Log struct {
	env      string
	telegram model.Telegram
}

func New(env string, telegram model.Telegram) Log {
	return Log{
		env:      env,
		telegram: telegram,
	}
}

func (l *Log) ConsoleLog(requestID, value any) {
	os.Stdout.WriteString(fmt.Sprintf("[cache-middleware-plugin] [%s] %v\n", requestID, value))
}

func (l *Log) TelegramLog(requestID, value any) {
	if l.telegram.Token != "" && l.telegram.ChatID != "" {
		params := url.Values{}
		params.Add("chat_id", l.telegram.ChatID)
		params.Add("text", fmt.Sprintf("[%s][cache-middleware-plugin]\nRequestID: %s\n%v", l.env, requestID, value))
		params.Add("parse_mode", "HTML")

		rs, errGet := http.Get(fmt.Sprintf("https://api.telegram.org/%s/sendMessage?%s", l.telegram.Token, params.Encode()))
		if errGet != nil {
			l.ConsoleLog(requestID, errGet.Error())
		}

		if rs.StatusCode != 200 {
			body, errRead := ioutil.ReadAll(rs.Body)
			if errRead != nil {
				l.ConsoleLog(requestID, errRead.Error())
			}

			rs.Body.Close()
			l.ConsoleLog(requestID, string(body))
		}
	}

	l.ConsoleLog(requestID, value)
}
