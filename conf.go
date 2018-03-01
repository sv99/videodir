package main

import (
	"io/ioutil"
	"github.com/kataras/iris"
	"github.com/sv99/htpasswd"
	"github.com/BurntSushi/toml"
	"github.com/dgrijalva/jwt-go"
)

func (conf *Config) Init(app *iris.Application)  {
	if _, err := toml.DecodeFile("./videodir.conf", &conf); err != nil {
		app.Logger().Warn("Config problems: " + err.Error())
	}
	// set LogLevel from config
	app.Logger().SetLevel(conf.LogLevel)

	// read private key
	app.Logger().Debug("key: ", conf.Key)
	signBytes, err := ioutil.ReadFile(conf.Key)
	if err != nil {
		app.Logger().Fatal("read key file error: ", err.Error())
		return
	}
	conf.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		app.Logger().Fatal("init private key error: ", err.Error())
		return
	}

	// read public key from certificate
	app.Logger().Debug("cert: ", conf.Cert)
	certBytes, err := ioutil.ReadFile(conf.Cert)
	if err != nil {
		app.Logger().Fatal("read certificate file error: ", err.Error())
		return
	}
	conf.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(certBytes)
	if err != nil {
		app.Logger().Fatal("init public key from certificate error: ", err.Error())
		return
	}

	// read .htpasswd
	conf.passwords, err = htpasswd.ParseHtpasswdFile(HTPASSWD)
	if err != nil {
		app.Logger().Fatal("read htpasswd error: ", err.Error())
		return
	}
}
