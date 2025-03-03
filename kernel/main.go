// LianDi - 链滴笔记，连接点滴
// Copyright (c) 2020-present, b3log.org
//
// LianDi is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package main

import (
	"encoding/json"
	"math/rand"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/liandi/kernel/cmd"
	"github.com/88250/liandi/kernel/model"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	model.InitLog()
	model.InitSessions()
	model.InitConf()
	model.InitMount()
	model.InitIndex()

	model.InitProcess()
	go model.ParentExited()
	model.CheckUpdatePeriodically()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/allocs", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/block", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/goroutine", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/heap", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/mutex", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/threadcreate", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))

	m := melody.New()
	m.Config.MaxMessageSize = 1024 * 1024 * 2
	r.GET("/ws", func(c *gin.Context) {
		if err := m.HandleRequest(c.Writer, c.Request); nil != err {
			model.Logger.Errorf("处理命令失败：%s", err)
		}
	})

	r.POST("/upload", model.Upload)
	r.POST("/upload/fetch", model.UploadFetch)

	m.HandleConnect(func(s *melody.Session) {
		model.AddPushChan(s)
		//sessionId, _ := s.Get("id")
		//model.Logger.Debugf("会话 [%s] 已连接", sessionId)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		model.RemovePushChan(s)
		//sessionId, _ := s.Get("id")
		//model.Logger.Debugf("会话 [%s] 已断开", sessionId)
	})

	m.HandleError(func(s *melody.Session, err error) {
		//sessionId, _ := s.Get("id")
		//model.Logger.Debugf("会话 [%s] 报错：%s", sessionId, err)
	})

	m.HandleClose(func(s *melody.Session, i int, str string) error {
		//sessionId, _ := s.Get("id")
		//model.Logger.Debugf("会话 [%s] 关闭：%v, %v", sessionId, i, str)
		return nil
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		model.Logger.Debugf("request [%s]", shortReqMsg(msg))
		request := map[string]interface{}{}
		if err := json.Unmarshal(msg, &request); nil != err {
			result := model.NewResult()
			result.Code = -1
			result.Msg = "Bad Request"
			responseData, _ := json.Marshal(result)
			s.Write(responseData)
			return
		}

		cmdStr := request["cmd"].(string)
		cmdId := request["reqId"].(float64)
		param := request["param"].(map[string]interface{})
		command := cmd.NewCommand(cmdStr, cmdId, param, s)
		if nil == command {
			result := model.NewResult()
			result.Code = -1
			result.Msg = "查找命令 [" + cmdStr + "] 失败"
			s.Write(result.Bytes())
			return
		}
		cmd.Exec(command)
	})

	handleSignal()

	addr := "127.0.0.1:" + model.ServerPort
	model.Logger.Infof("内核进程 [v%s] 正在启动，监听端口 [%s]", model.Ver, "http://"+addr)
	if err := r.Run(addr); nil != err {
		model.Logger.Errorf("启动链滴笔记内核失败 [%s]", err)
	}
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		s := <-c
		model.Logger.Infof("收到系统信号 [%s]，退出内核进程", s)

		model.Close()
		os.Exit(0)
	}()
}

func shortReqMsg(msg []byte) []byte {
	s := gulu.Str.FromBytes(msg)
	max := 128
	if len(s) > max {
		count := 0
		for i := range s {
			count++
			if count > max {
				return gulu.Str.ToBytes(s[:i] + "...")
			}
		}
	}
	return msg
}
