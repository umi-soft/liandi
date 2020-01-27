// LianDi - 链滴笔记，链接点滴
// Copyright (c) 2020-present, b3log.org
//
// Lute is licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.

package util

import (
	"github.com/88250/gowebdav"
	"github.com/88250/gulu"
	"github.com/blevesearch/bleve"

	_ "github.com/blevesearch/bleve/analysis/lang/cjk"
)

var index bleve.Index

func InitSearch() {
	var err error

	if gulu.File.IsExist(IndexPath) {
		index, err = bleve.Open(IndexPath)
		if nil != err {
			logger.Fatalf("加载搜索索引失败：%s", err)
			return
		}
	} else {
		mapping := bleve.NewIndexMapping()
		mapping.DefaultAnalyzer = "cjk"
		index, err = bleve.New(IndexPath, mapping)
		if nil != err {
			logger.Fatalf("创建搜索索引失败：%s", err)
			return
		}
	}

	go func() {
		for _, dir := range Conf.Dirs {
			files := dir.Files("/")
			for _, file := range files {
				content, err := dir.Get(file.(gowebdav.File).Path())
				if nil == err {
					Index(content)
				}
			}
		}
	}()
}

func Index(content string) {
	id := gulu.Rand.String(7)
	if err := index.Index(id, content); nil != err {
		logger.Errorf("索引失败：%s", err)
	}
}

func Search(text string) {
	query := bleve.NewMatchQuery(text)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if nil != err {
		logger.Warnf("搜索失败：%s", err)
		return
	}
	logger.Infof("%+v", searchResults)
}
