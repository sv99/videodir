package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kataras/iris/httptest"
)

const videoTest = "./videoTest/"
const videoTest1 = "./videoTest1/"
const cam01 = "cam01"
const cam02 = "cam02"
const videoFile = "record"

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return err != nil
}

func createCamFolder(path,cam string, first int) {
	p := filepath.Join(path, cam)
	os.MkdirAll(p, os.ModePerm)

	for i := 0; i < 5; i++{
		file, err := os.Create(filepath.Join(p, fmt.Sprintf("%s%02d.rec", videoFile, i + first)))
		if isError(err) {
			return
		}
		_, err = file.WriteString("12345")
		if isError(err) {
			file.Close()
			return
		}
		file.Close()
	}
}


func createVideoTestFolder(path string, first int) {
	createCamFolder(path, cam01, first)
	createCamFolder(path, cam02, first)
}

func cleanTestFolder(path string) {
	os.RemoveAll(path)
}

func TestVideoDir(t *testing.T) {
	// create test volumes
	createVideoTestFolder(videoTest, 1)
	defer cleanTestFolder(videoTest)
	createVideoTestFolder(videoTest1, 6)
	defer cleanTestFolder(videoTest1)

	// test config with test volumes
	conf := DefaultConfig()
	conf.TOML("./videodir_test.conf")
	app := newApp(&conf)
	e := httptest.New(t, app)

	// home without auth
	e.GET("/").Expect().Status(httptest.StatusOK)

	// api without auth
	e.GET("/api/v1/version").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/api/v1/volumes").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/api/v1/list").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/api/v1/file").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/api/v1/filesize").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/api/v1/remove").Expect().Status(httptest.StatusUnauthorized)

	user := UserCredentials{
		Username: "dima",
		Password: "dima_dima_dima",
	}

	userInvalid := UserCredentials{
		Username: "dima",
		Password: "ku_ku",
	}

	// get jwt token
	e.POST("/login").WithJSON(userInvalid).Expect().Status(httptest.StatusUnauthorized)
	resp := e.POST("/login").WithJSON(user).Expect().Status(httptest.StatusOK)
	resp.JSON().Object().ContainsKey("token")
	token := resp.JSON().Object().Value("token").String().Raw()

	// check access with JWT token
	resp = e.GET("/api/v1/version").WithHeader("Authorization", "Bearer " + token).Expect()
	resp.Status(httptest.StatusOK)
	resp.JSON().Object().ValueEqual("version", VERSION)

	// get list volumes
	resp = e.GET("/api/v1/volumes").WithHeader("Authorization", "Bearer " + token).Expect()
	resp.Status(httptest.StatusOK)
	resp.JSON().Array().Elements(videoTest, videoTest1)

	emptyPath := Path{
		Path: []string{""},
	}

	filePath := Path{
		Path: []string{cam01, videoFile + "04.rec"},
	}

	// list all folders from all volumes
	resp = e.POST("/api/v1/list").WithHeader("Authorization", "Bearer " + token).WithJSON(emptyPath).Expect()
	resp.Status(httptest.StatusOK)
	resp.JSON().Array().Elements(cam01, cam02, cam01, cam02)

	// check file size
	resp = e.POST("/api/v1/filesize").WithHeader("Authorization", "Bearer " + token).WithJSON(filePath).Expect()
	resp.Status(httptest.StatusOK)
	resp.JSON().Object().ValueEqual("size", 5)

	// check file contents
	resp = e.POST("/api/v1/file").WithHeader("Authorization", "Bearer " + token).WithJSON(filePath).Expect()
	resp.Status(httptest.StatusOK)
	res := resp.Body().Raw()
	if res != "12345" {
		t.Error("Get file error!")
	}

	// delete file
	resp = e.POST("/api/v1/remove").WithHeader("Authorization", "Bearer " + token).WithJSON(filePath).Expect()
	resp.Status(httptest.StatusOK)

	// file deleted return size = 0
	resp = e.POST("/api/v1/filesize").WithHeader("Authorization", "Bearer " + token).WithJSON(filePath).Expect()
	resp.Status(httptest.StatusOK)
	resp.JSON().Object().ValueEqual("size", 0)

}
