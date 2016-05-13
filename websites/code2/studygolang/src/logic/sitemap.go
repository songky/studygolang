// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package logic

import (
	"os"
	"strconv"
	"text/template"
	"time"
	"util"

	"github.com/polaris1119/config"
	"github.com/polaris1119/logger"

	. "db"
	"model"
)

// 自定义模板函数
var funcMap = template.FuncMap{
	"time_format": func(s string) string {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		if err != nil {
			return s
		}

		return t.Format(time.RFC3339)
	},
}

var sitemapTpl = template.Must(template.New("sitemap.xml").Funcs(funcMap).ParseFiles(config.TemplateDir + "/sitemap.xml"))
var sitemapIndexTpl = template.Must(template.ParseFiles(config.TemplateDir + "/sitemapindex.xml"))

var sitemapPath = config.ROOT + "/sitemap/"

func init() {
	if !util.Exist(sitemapPath) {
		err := os.MkdirAll(sitemapPath, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func GenSitemap() {
	sitemapFiles := []string{}

	// 首页
	home := map[string]string{
		"loc":      "http://" + config.ConfigFile.MustValue("global", "domain"),
		"lastmode": time.Now().Format(time.RFC3339),
	}

	var (
		little = 1
		step   = 4999
		large  = little + step
	)

	// 文章
	var (
		articles = make([]*model.Article, 0)
		err      error
	)
	for {
		sitemapFile := "sitemap_article_" + strconv.Itoa(large) + ".xml"

		err = MasterDB.Where("id BETWEEN ? AND ? AND status!=?", little, large, model.ArticleStatusOffline).Select("id,mtime").Find(&articles)
		little = large + 1
		large = little + step

		if err != nil {
			continue
		}

		if len(articles) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":     home,
			"articles": articles,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}
	}

	little = 1
	large = little + step

	// 主题（帖子）
	topics := make([]*model.Topic, 0)
	for {
		sitemapFile := "sitemap_topic_" + strconv.Itoa(large) + ".xml"

		err = MasterDB.Where("tid BETWEEN ? AND ? AND flag IN(?,?)", little, large, 0, 1).Select("tid,mtime").Find(&topics)
		little, large = large+1, little+step

		if err != nil {
			continue
		}

		if len(topics) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":   home,
			"topics": topics,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}
	}

	little = 1
	large = little + step

	// 资源
	resources := make([]*model.Resource, 0)
	for {
		sitemapFile := "sitemap_resource_" + strconv.Itoa(large) + ".xml"

		err = MasterDB.Where("id BETWEEN ? AND ?", little, large).Select("id,mtime").Find(&resources)
		little, large = large+1, little+step

		if err != nil {
			continue
		}

		if len(resources) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":      home,
			"resources": resources,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}
	}

	little = 1
	large = little + step

	// 项目
	projects := make([]*model.OpenProject, 0)
	for {
		sitemapFile := "sitemap_project_" + strconv.Itoa(large) + ".xml"

		err = MasterDB.Where("id BETWEEN ? AND ?", little, large).Select("id,uri,mtime").Find(&resources)
		little, large = large+1, little+step

		if err != nil {
			continue
		}

		if len(projects) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":     home,
			"projects": projects,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}
	}

	little = 1
	large = little + step

	// wiki
	// wiki := model.NewWiki()
	// for {
	// 	sitemapFile := "sitemap_wiki_" + strconv.Itoa(large) + ".xml"

	// 	wikis, err := wiki.Where("id BETWEEN ? AND ?", little, large).FindAll("id", "uri", "mtime")
	// 	little, large = large+1, little+step

	// 	if err != nil {
	// 		continue
	// 	}

	// 	if len(wikis) == 0 {
	// 		break
	// 	}

	// 	data := map[string]interface{}{
	// 		"home":  home,
	// 		"wikis": wikis,
	// 	}

	// 	if err = output(sitemapFile, data); err == nil {
	// 		sitemapFiles = append(sitemapFiles, sitemapFile)
	// 	}
	// }

	file, err := os.Create(sitemapPath + "sitemapindex.xml")
	if err != nil {
		logger.Errorln("gen sitemap index file error:", err)
		return
	}
	defer file.Close()

	err = sitemapIndexTpl.Execute(file, map[string]interface{}{
		"home":         home,
		"sitemapFiles": sitemapFiles,
	})
	if err != nil {
		logger.Errorln("execute sitemap index template error:", err)
	}
}

func output(filename string, data map[string]interface{}) (err error) {
	var file *os.File
	file, err = os.Create(sitemapPath + filename)
	if err != nil {
		logger.Errorln("open file error:", err)
		return
	}
	defer file.Close()
	if err = sitemapTpl.Execute(file, data); err != nil {
		logger.Errorln("execute template error:", err)
	}

	return
}
