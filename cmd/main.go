package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Page struct {
	Title     string
	Committer []string
	Password  []string
	Name      []string
}

func (p *Page) mail() error {
	host := "mail.narlabs.org.tw"
	port := 465

	auth := LoginAuth(mail_account, mail_password)
	c, err := CreateSMTPClient(fmt.Sprintf("%s:%d", host, port), auth)
	if err != nil {
		return err
	}
	defer c.Close()

	for i := 0; i < 9; i++ {

		v := p.Committer[i]
		n := p.Name[i]

		glog.Infof("%d,%s", i+1, v)

		toEmail := v
		header := make(map[string]string)
		header["From"] = "工作小組" + "<" + mail_account + ">"
		header["To"] = toEmail
		header["Subject"] = "委員編號"
		header["Content-Type"] = "text/html; charset=UTF-8"
		body := fmt.Sprintf("%s 委員好: <br>您抽到的號碼為 %d。<br><br>工作小組", n, i+1)
		message := ""
		for k, v := range header {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\n" + body

		err := SendMailUsingTLS(
			c,
			mail_account,
			[]string{toEmail},
			[]byte(message),
		)
		if err != nil {
			glog.Warningf("Send to %s fail: %s", toEmail, err.Error())
		}
	}

	return nil
}

func (p *Page) save() error {

	for _, v := range p.Password {
		glog.Info(v)
	}

	filename := p.Title + ".csv"

	file, err := os.Create(filename)
	if err != nil {
		glog.Infof("open file is failed, err: %s", err.Error())
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Write([]string{"number", "mail"})
	for i, v := range p.Committer {
		r := []string{fmt.Sprintf("%d", i+1), v}
		w.Write(r)
	}

	w.Flush()

	fn := fmt.Sprintf("/tmp/%s.zip", p.Title)
	e := Zip(fn, p.Title+".csv", p.Password[0]+p.Password[1])
	if e != nil {
		glog.Info(e.Error())
	}

	return nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := Page{Title: title}
	renderTemplate(w, "view", &p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {

	title := fmt.Sprintf("%d", time.Now().Unix())
	p := &Page{Title: title}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: title, Committer: []string{}}

	for i := 0; i < 9; i++ {
		c := r.FormValue(fmt.Sprintf("committer-%d", i+1))
		n := r.FormValue(fmt.Sprintf("name-%d", i+1))
		p.Committer = append(p.Committer, c)
		p.Name = append(p.Name, n)
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

		y := p.Name[i]
		p.Name[i] = p.Name[n]
		p.Name[n] = y
	}

	err := p.mail()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func downloadHandler(w http.ResponseWriter, r *http.Request, title string) {

	file := fmt.Sprintf("/tmp/%s.zip", title)

	nn := fmt.Sprintf("%s.zip", title)
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

var validPath = regexp.MustCompile("^/(save|view|download)/([a-zA-Z0-9]+)$")

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

var mail_password string
var mail_account string

func main() {

	flag.StringVar(&mail_password, "password", "1234567", "smtp password")
	flag.StringVar(&mail_account, "mail", "abc@test.com", "smtp mail account")

	flag.Parse()
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/", editHandler)
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/download/", makeHandler(downloadHandler))
	glog.Info(http.ListenAndServe(":8080", nil))
}
