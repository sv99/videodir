package main

import (
	"crypto/rsa"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/dgrijalva/jwt-go"
	"github.com/foomo/htpasswd"
	"github.com/kataras/iris/v12"
	"io/ioutil"
)

type Config struct {
	LogLevel   string
	ServerAddr string
	Cert       string
	Key        string
	VideoDirs  []string

	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	passwords htpasswd.HashedPasswords
}

func DefaultConfig() Config {
	return Config{
		LogLevel:   "info",
		ServerAddr: ":8443",
		Cert:       "server.crt",
		Key:        "server.key",
	}
}

func (conf *Config) TOML(fpath string) {
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		panic(fmt.Sprintf("Config problems: %s", err.Error()))
	}
}

func (conf *Config) Init(app *iris.Application)  {
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
