// LianDi - 链滴笔记，连接点滴
// Copyright (c) 2020-present, b3log.org
//
// LianDi is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package model

import (
	"strings"
	"unicode/utf8"
)

func InitIndex() {
	for _, dir := range Conf.Dirs {
		go dir.Index()
	}
}

var (
	// docs 用于维护所有已挂载的文档。
	docs []*Doc
)

type Doc struct {
	URL     string
	Path    string
	Content string
}

type Snippet struct {
	Dir     *Dir   `json:"dir"`
	Path    string `json:"path"`
	Ln      int    `json:"ln"`
	Col     int    `json:"col"`
	Index   int    `json:"index"`
	Content string `json:"content"`
}

func (dir *Dir) RemoveIndexDocDir(dirPath string) {
	for i := 0; i < len(docs); i++ {
		if dir.URL == docs[i].URL && strings.HasPrefix(docs[i].Path, dirPath) {
			docs = append(docs[:i], docs[i+1:]...)
			i--
		}
	}
}

func (dir *Dir) MoveIndexDocsDir(dirPath, newDirPath string) {
	for _, d := range docs {
		if dir.URL == d.URL && strings.HasPrefix(d.Path, dirPath) {
			d.Path = strings.Replace(d.Path, dirPath, newDirPath, -1)
		}
	}
}

func (dir *Dir) MoveIndexDoc(path, newPath string) {
	for _, d := range docs {
		if dir.URL == d.URL && path == d.Path {
			d.Path = newPath
			break
		}
	}
}

func (dir *Dir) RemoveIndexDoc(path string) {
	for i, doc := range docs {
		if doc.URL == dir.URL && doc.Path == path {
			docs = docs[:i+copy(docs[i:], docs[i+1:])]
			break
		}
	}
}

func (dir *Dir) IndexDoc(path, content string) {
	doc := &Doc{URL: dir.URL, Path: path, Content: content}
	for i, d := range docs {
		if doc.URL == d.URL && doc.Path == d.Path {
			docs = docs[:i+copy(docs[i:], docs[i+1:])]
			break
		}
	}
	docs = append(docs, doc)
}

func Search(keyword string) (ret []*Snippet) {
	ret = []*Snippet{}
	for _, doc := range docs {
		snippets := searchDoc(keyword, doc)
		ret = append(ret, snippets...)
	}
	return
}

func searchDoc(keyword string, doc *Doc) (ret []*Snippet) {
	lines := strings.Split(doc.Content, "\n")
	index := 0
	maxPartLen := 32
	for idx, line := range lines {
		if pos := strings.Index(strings.ToLower(line), strings.ToLower(keyword)); -1 != pos {
			var before []rune
			var count int
			for i := pos; 0 < i; { // 关键字前面太长的话缩短一些
				r, size := utf8.DecodeLastRuneInString(line[:i])
				i -= size
				before = append([]rune{r}, before...)
				count++
				if maxPartLen < count {
					break
				}
			}

			highlight := string(before) + "<mark>" + line[pos:pos+len(keyword)] + "</mark>" + line[pos+len(keyword):]
			dir := Conf.Dir(doc.URL)
			snippet := &Snippet{Dir: dir,
				Path: doc.Path,
				Ln:   idx + 1, Col: pos + 1, Index: index,
				Content: highlight}
			ret = append(ret, snippet)
			index++
		}
	}
	return ret
}
