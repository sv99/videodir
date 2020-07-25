package videodir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber"
	"github.com/gofiber/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var validUser = UserCredentials{
	Username: "dima",
	Password: "dima_dima_dima",
}

var invalidUser = UserCredentials{
	Username: "dima",
	Password: "ku_ku",
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return err != nil
}

func checkRoute(t *testing.T, app *fiber.App, methode string,
	target string, body io.Reader, statusCode int) {
	resp, err := app.Test(httptest.NewRequest(methode, target, body))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, statusCode, resp.StatusCode)
	utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
}

func testApp() *AppServer {
	// test config with test volumes
	conf := DefaultConfig()
	err := conf.TOML("./videodir_test.conf")
	if err != nil {
		log.Panicf("Config load problems: %s", err.Error())
	}
	app := NewApp(&conf)
	return app
}

func TestRoutesNoAuth(t *testing.T) {
	app := testApp()

	// home without auth
	checkRoute(t, app.App, "GET", "/", nil, fiber.StatusOK)

	// api without auth
	checkRoute(t, app.App, "GET", "/api/v1/version", nil, fiber.StatusUnauthorized)
	checkRoute(t, app.App, "GET", "/api/v1/volumes", nil, fiber.StatusUnauthorized)
	checkRoute(t, app.App, "POST", "/api/v1/list", nil, fiber.StatusUnauthorized)
	checkRoute(t, app.App, "POST", "/api/v1/file", nil, fiber.StatusUnauthorized)
	checkRoute(t, app.App, "POST", "/api/v1/filesize", nil, fiber.StatusUnauthorized)
	checkRoute(t, app.App, "POST", "/api/v1/remove", nil, fiber.StatusUnauthorized)
}

func getToken(t *testing.T, app *fiber.App) string {
	// get jwt token ok user
	rawUser, err := json.Marshal(validUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/login", bytes.NewReader(rawUser)))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	var token Token
	err = decoder.Decode(&token)
	utils.AssertEqual(t, nil, err)
	return token.Token
}

func TestRoutesLogin(t *testing.T) {
	app := testApp()

	// get jwt token bad user
	rawUser, err := json.Marshal(invalidUser)
	resp, err := app.App.Test(httptest.NewRequest("POST", "/login", bytes.NewReader(rawUser)))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusUnauthorized, resp.StatusCode)

	// get jwt token ok user
	_ = getToken(t, app.App)
}

func getAuthRequest(methode string, target string, body io.Reader, token string) *http.Request {
	req := httptest.NewRequest(methode, target, body)
	req.Header.Set("Authorization", "Bearer " + token)
	return req
}

const videoTest = "./videoTest/"
const videoTest1 = "./videoTest1/"

func TestRoutesWirthAuth(t *testing.T) {
	app := testApp()

	token := getToken(t, app.App)
	req := getAuthRequest("GET", "/api/v1/version", nil, token)
	resp, err := app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	req = getAuthRequest("GET", "/api/v1/volumes", nil, token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	var volumes []string
	err = decoder.Decode(&volumes)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 2, len(volumes))
	utils.AssertEqual(t, videoTest, volumes[0])
	utils.AssertEqual(t, videoTest1, volumes[1])
}

// check access to the file system
const cam01 = "cam01"
const cam02 = "cam02"
const videoFile = "record"

func createCamFolder(path, cam string, first int, count int) {
	p := filepath.Join(path, cam)
	_ = os.MkdirAll(p, os.ModePerm)

	for i := 0; i < count; i++ {
		file, err := os.Create(filepath.Join(p, fmt.Sprintf("%s%02d.rec", videoFile, i + first)))
		if isError(err) {
			return
		}
		_, err = file.WriteString("12345")
		if isError(err) {
			_ = file.Close()
			return
		}
		_ = file.Close()
	}
}

func createVideoTestFolder(path string, first int, count int) {
	createCamFolder(path, cam01, first, count)
	createCamFolder(path, cam02, first, count)
}

func cleanTestFolder(path string) {
	_ = os.RemoveAll(path)
}

func TestList(t *testing.T) {
	// create test volumes
	createVideoTestFolder(videoTest, 1, 5)
	defer cleanTestFolder(videoTest)
	createVideoTestFolder(videoTest1, 6, 5)
	defer cleanTestFolder(videoTest1)

	app := testApp()
	token := getToken(t, app.App)

	var path Path
	rawRoot, err := json.Marshal(path)
	req := getAuthRequest("POST", "/api/v1/list", bytes.NewReader(rawRoot), token)
	resp, err := app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	var files []string
	err = decoder.Decode(&files)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 4, len(files))
	utils.AssertEqual(t, []string{"cam01", "cam02", "cam01", "cam02"}, files)

	// list path cam01/
	path.Path = []string{"cam01"}
	rawCam01, err := json.Marshal(path)
	req = getAuthRequest("POST", "/api/v1/list", bytes.NewReader(rawCam01), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&files)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 10, len(files))
	utils.AssertEqual(t, []string{
		"record01.rec", "record02.rec", "record03.rec", "record04.rec", "record05.rec",
		"record06.rec", "record07.rec", "record08.rec", "record09.rec", "record10.rec",
	}, files)
}

func TestGetFile(t *testing.T) {
	// create test volumes
	createVideoTestFolder(videoTest, 1, 5)
	defer cleanTestFolder(videoTest)

	app := testApp()
	token := getToken(t, app.App)

	// get file size
	var path Path
	path.Path = []string{"cam01", "record01.rec"}
	rawRecord01, err := json.Marshal(path)
	path.Path = []string{"cam01", "record10.rec"}
	rawRecord10, err := json.Marshal(path)
	req := getAuthRequest("POST", "/api/v1/filesize", bytes.NewReader(rawRecord01), token)
	resp, err := app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	var filesize map[string]int
	err = decoder.Decode(&filesize)
	utils.AssertEqual(t, nil, err)
	fmt.Println(filesize)
	utils.AssertEqual(t, 5, filesize["size"])

    // get not exists file size
	req = getAuthRequest("POST", "/api/v1/filesize", bytes.NewReader(rawRecord10), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusBadRequest, resp.StatusCode)

	// get file
	req = getAuthRequest("POST", "/api/v1/file", bytes.NewReader(rawRecord01), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	file, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 5, len(file))
	utils.AssertEqual(t, "12345", string(file))

	// get not exists file
	req = getAuthRequest("POST", "/api/v1/file", bytes.NewReader(rawRecord10), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRemove(t *testing.T) {
	// create test volumes
	createVideoTestFolder(videoTest, 1, 2)
	defer cleanTestFolder(videoTest)
	createVideoTestFolder(videoTest1, 3, 4)
	defer cleanTestFolder(videoTest1)

	app := testApp()
	token := getToken(t, app.App)

	var path Path
	rawRoot, err := json.Marshal(path)
	path.Path = []string{"cam01"}
	rawCam01, err := json.Marshal(path)
	path.Path = []string{"cam01", "record01.rec"}
	rawRecord01, err := json.Marshal(path)
	// remove file
	req := getAuthRequest("POST", "/api/v1/remove", bytes.NewReader(rawRecord01), token)
	resp, err := app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	// list path cam01/ after delete
	req = getAuthRequest("POST", "/api/v1/list", bytes.NewReader(rawCam01), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	var files []string
	err = decoder.Decode(&files)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 5, len(files))
	utils.AssertEqual(t, []string{
		"record02.rec", "record03.rec", "record04.rec", "record05.rec", "record06.rec",
	}, files)

	// remove not empty dir
	req = getAuthRequest("POST", "/api/v1/remove", bytes.NewReader(rawCam01), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

	// list path cam01/ after delete dir
	req = getAuthRequest("POST", "/api/v1/list", bytes.NewReader(rawCam01), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&files)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 0, len(files))

	// list root after delete directory
	req = getAuthRequest("POST", "/api/v1/list", bytes.NewReader(rawRoot), token)
	resp, err = app.App.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&files)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, 2, len(files))
	utils.AssertEqual(t, []string{
		"cam02", "cam02",
	}, files)

}
