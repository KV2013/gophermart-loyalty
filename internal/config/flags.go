package config

import (
	flag "github.com/spf13/pflag"
)

var flagServerAddr string
var flagBaseURL string
var flagLogLevel string
var flagDatabaseURI string
var flagJWTSecretKey string
var flagAccrualSystemAddress string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagServerAddr, "address", "a", "", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "базовый адрес сервиса лояльности")
	flag.StringVarP(&flagLogLevel, "loglevel", "l", "", "log level")
	flag.StringVarP(&flagDatabaseURI, "database-uri", "d", "", "database URI")
	flag.StringVarP(&flagJWTSecretKey, "jwt-secret", "j", "", "JWT signing secret key")
	flag.StringVarP(&flagAccrualSystemAddress, "accrual-system-address", "r", "", "адрес API сервиса расчёта начислений")

	flag.Parse()
}
