// LianDi - 链滴笔记，连接点滴
// Copyright (c) 2020-present, b3log.org
//
// LianDi is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package cmd

import (
	"github.com/88250/liandi/kernel/model"
	"path"
)

type mkdir struct {
	*BaseCmd
}

func (cmd *mkdir) Exec() {
	ret := model.NewCmdResult(cmd.Name(), cmd.id)
	url := cmd.param["url"].(string)
	url = model.NormalizeURL(url)
	p := cmd.param["path"].(string)
	err := model.Mkdir(url, p)
	if nil != err {
		ret.Code = -1
		ret.Msg = err.Error()
		cmd.Push(ret.Bytes())
		return
	}

	dir := model.Conf.Dir(url)
	files, _ := model.Ls(url, path.Dir(p))
	ret.Data = map[string]interface{}{
		"dir":  dir,
		"files": files,
		"name": path.Base(p),
	}
	cmd.Push(ret.Bytes())
}

func (cmd *mkdir) Name() string {
	return "mkdir"
}
