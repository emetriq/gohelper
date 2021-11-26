package log

import (
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
)

func InitGraylog(ip, port, facility string) {
	if ip != "" && port != "" && facility != "" {
		hook := graylog.NewGraylogHook(ip+":"+port, map[string]interface{}{"facility": facility})
		Logger.AddHook(hook)
		Logger.Debug("Logging on Graylog enabled")
	} else {
		Logger.Debug("Logging on Graylog disabled")
	}
}
