// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"archive/zip"
	"encoding/csv"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Page struct {
	Title     string
	Committer []string
	Password  []string
}

func (p *Page) save() error {

	// todo: send mail

	for _, v := range p.Password {
		log.Printf(v)
	}

	for i, v := range p.Committer {
		log.Printf("%d,%s", i+1, v)
	}

	filename := p.Title + ".csv"

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("open file is failed, err: ", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Write([]string{"number", "mail"})
	for i, v := range p.Committer {
		r := []string{fmt.Sprintf("%d", i+1), v}
		w.Write(r)
	}

	w.Flush()

	return nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := Page{Title: title}
	renderTemplate(w, "view", &p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {

	p := &Page{Title: title}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: title, Committer: []string{}}

	for i := 0; i < 9; i++ {
		c := r.FormValue(fmt.Sprintf("committer-%d", i+1))
		p.Committer = append(p.Committer, c)
	}
	for i := 0; i < 2; i++ {
		c := r.FormValue(fmt.Sprintf("password-%d", i+1))
		p.Password = append(p.Password, c)
	}

	l := len(p.Committer) - 1
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i <= l; i++ {
		n := rand.Intn(l)
		// swap
		x := p.Committer[i]
		p.Committer[i] = p.Committer[n]
		p.Committer[n] = x
	}

	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func downloadHandler(w http.ResponseWriter, r *http.Request, title string) {

	ts := time.Now().Unix()
	file := fmt.Sprintf("/tmp/%s-%d.zip", title, ts)

	e := Zip(file, title+".csv")
	if e != nil {
		fmt.Printf(e.Error())
	}

	nn := fmt.Sprintf("%s-%d.zip", title, ts)
	w.Header().Set("Content-Disposition", "attachment; filename="+nn)
	w.Header().Set("Content-Type", "application/zip")

	http.ServeFile(w, r, file)
}

var templates = template.Must(template.ParseFiles("template/edit.html", "template/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|download)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/download/", makeHandler(downloadHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Zip(dst, src string) (err error) {
	// 创建准备写入的文件
	fw, err := os.Create(dst)
	defer fw.Close()
	if err != nil {
		return err
	}

	// 通过 fw 来创建 zip.Write
	zw := zip.NewWriter(fw)
	defer func() {
		// 检测一下是否成功关闭
		if err := zw.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(path, string(filepath.Separator))

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return
		}

		// 将打开的文件 Copy 到 w
		_, err = io.Copy(w, fr)
		if err != nil {
			return
		}
		// 输出压缩的内容
		log.Printf("Create zip file: %s", dst)

		return nil
	})
}
