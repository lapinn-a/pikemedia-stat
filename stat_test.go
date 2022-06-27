package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	stat   *Stat
	router *gin.Engine
}

func (s *TestSuite) SetupSuite() {
	db, err := sql.Open("sqlite3", "file:test1?mode=memory&cache=shared")
	if err != nil {
		log.Fatalf("FATAL: Error opening database: %s\n", err)
	}
	startTime := time.Now()
	stat := NewStat(db, startTime)

	err = stat.RunMigrations()
	if err != nil {
		log.Fatalf("FATAL: Error running migrations: %s\n", err)
	}
	s.router = stat.Router()
}

func TestStatSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestPing() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), "{\"status\":\"up\"}", w.Body.String())
}

func (s *TestSuite) TestCollect() {
	body := `[{"viewerId":10366,"name":"Роман","lastName":"XXXXX","isChatName":false,"email":"aaaa@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:37:24+03:00","leaveTime":"2021-07-30T15:45:43+03:00","spentTime":461000000000,"spentTimeDeltaPercent":14,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"62.152.34.188","platform":"OS X 10.15.7 64-bit","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1440x900","screenData_resolution":"1440x900"}},{"viewerId":11181,"name":"Сергей","lastName":"Сергеев","isChatName":false,"email":"bbbbb@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T14:12:48+03:00","leaveTime":"2021-07-30T14:25:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.137.131.4","platform":"Windows 10 64-bit","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1920x1040","screenData_resolution":"1920x1080"}},{"viewerId":11281,"name":"Василий","lastName":"Александров","isChatName":false,"email":"xxxxx@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T14:20:48+03:00","leaveTime":"2021-07-30T15:40:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.197.131.4","platform":"Windows 7 64-bit","browserClient":"Chrome 92.0.4515.100","screenData_viewPort":"1280x720","screenData_resolution":"1280x720"}},{"viewerId":14281,"name":"Александр","lastName":"Васильев","isChatName":false,"email":"zzzzz@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:39:48+03:00","leaveTime":"2021-07-30T15:50:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.197.136.4","platform":"Windows 7 64-bit","browserClient":"Firefox 15.10","screenData_viewPort":"1280x700","screenData_resolution":"1280x700"}}]`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), "{\"result\":\"success\"}", w.Body.String())
}

func (s *TestSuite) TestCollectConstraint() {
	body := `[{"viewerId":10366,"name":"Роман","lastName":"XXXXX","isChatName":false,"email":"aaaa@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:37:24+03:00","leaveTime":"2021-07-30T15:45:43+03:00","spentTime":461000000000,"spentTimeDeltaPercent":14,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"62.152.34.188","platform":"OS X 10.15.7 64-bit","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1440x900","screenData_resolution":"1440x900"}},{"viewerId":11181,"name":"Сергей","lastName":"Сергеев","isChatName":false,"email":"bbbbb@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T14:12:48+03:00","leaveTime":"2021-07-30T14:25:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.137.131.4","platform":"Windows 10 64-bit","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1920x1040","screenData_resolution":"1920x1080"}},{"viewerId":11281,"name":"Василий","lastName":"Александров","isChatName":false,"email":"xxxxx@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T14:20:48+03:00","leaveTime":"2021-07-30T15:40:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.197.131.4","platform":"Windows 7 64-bit","browserClient":"Chrome 92.0.4515.100","screenData_viewPort":"1280x720","screenData_resolution":"1280x720"}},{"viewerId":14281,"name":"Александр","lastName":"Васильев","isChatName":false,"email":"zzzzz@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:39:48+03:00","leaveTime":"2021-07-30T15:50:25+03:00","spentTime":676000000000,"spentTimeDeltaPercent":9,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"79.197.136.4","platform":"Windows 7 64-bit","browserClient":"Firefox 15.10","screenData_viewPort":"1280x700","screenData_resolution":"1280x700"}}]`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	assert.Equal(s.T(), "{\"result\":\"failed\"}", w.Body.String())
}

func (s *TestSuite) TestCollectIncorrectType() {
	// id is string
	body := `[{"viewerId":"10367","name":"Роман","lastName":"XXXXX","isChatName":false,"email":"aaaa@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:37:24+03:00","leaveTime":"2021-07-30T15:45:43+03:00","spentTime":461000000000,"spentTimeDeltaPercent":14,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"62.152.34.188","platform":"OS X 10.15.7 64-bit","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1440x900","screenData_resolution":"1440x900"}}]`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	assert.Equal(s.T(), "{\"result\":\"failed\"}", w.Body.String())
}

func (s *TestSuite) TestCollectNull() {
	// platform is null
	body := `[{"viewerId":"","name":"Роман","lastName":"XXXXX","isChatName":false,"email":"aaaa@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:37:24+03:00","leaveTime":"2021-07-30T15:45:43+03:00","spentTime":461000000000,"spentTimeDeltaPercent":14,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"62.152.34.188","platform":null,"browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1440x900","screenData_resolution":"1440x900"}}]`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	assert.Equal(s.T(), "{\"result\":\"failed\"}", w.Body.String())
}

func (s *TestSuite) TestCollectAbsence() {
	// lastName does not exists
	body := `[{"viewerId":12345,"name":"Роман","lastName":"XXXXX","isChatName":false,"email":"aaaa@pikemedia.ru","isChatEmail":false,"joinTime":"2021-07-30T15:37:24+03:00","leaveTime":"2021-07-30T15:45:43+03:00","spentTime":461000000000,"spentTimeDeltaPercent":14,"chatCommentsTotal":0,"chatCommentsDeltaPercent":0,"anotherFields":[],"browserClientInfo":{"userIP":"62.152.34.188","browserClient":"Chrome 92.0.4515.107","screenData_viewPort":"1440x900","screenData_resolution":"1440x900"}}]`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	assert.Equal(s.T(), "{\"result\":\"failed\"}", w.Body.String())
}

func (s *TestSuite) TestCollectEmpty() {
	// lastName does not exists
	body := ``
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/collect", strings.NewReader(body))
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	assert.Equal(s.T(), "{\"result\":\"failed\"}", w.Body.String())
}
