package main

import (
	"net/http"
	"text/template"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var config struct {
	MysqlLogin    string
	MysqlPassword string
	MysqlHost     string
	MysqlDb       string
}

type Page struct {
	Title  string
	Msg    string
	ContWr string
	Phone  string
	SIP1   string
	SIP2   string
}

type Phone struct {
	Name string
	Mac  string
	Ip   string
}
type SipUser struct {
	User        string `db:"internalnumber"`
	Description string `db:"description"`
	Password    string `db:"password"`
}

type User struct {
	UserLog   int
	UserPas   string
	UserName  string
	AnableReg bool
}

type PhoneConf struct {
	Users         [2]User
	VlanPhone     bool
	VlanPhoneNumb int
	VlanComp      bool
	VlanCompNumb  int
	Version       string
}

func (c *PhoneConf) MakeConfig(pf *Phone) (string, error) {
	latexTemplate, err := template.ParseFiles("TelConfig.xml")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + pf.Mac + ".xml")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	err = latexTemplate.ExecuteTemplate(outputLatexFile, "TelConfig.xml", c)
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles\\" + pf.Mac + ".xml"
	return pathToTexFile, nil
}

func loadConfig(path string) error {
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, &config)
}

var (
	configFile = flag.String("Config", "conf.json", "Where to read the Config from")
)

func index(w http.ResponseWriter, r *http.Request) {
	dataBase, err := sqlx.Connect("mysql", config.MysqlLogin+":"+config.MysqlPassword+"@tcp("+config.MysqlHost+")/"+config.MysqlDb+"?charset=utf8")
	defer dataBase.Close()
	if err != nil {
		log.Print(err)
	}
	PhoneData := make([]Phone, 0)
	err = dataBase.Select(&PhoneData, "SELECT `mac`, `name`, INET_NTOA(`ip`) AS ip FROM `unetmap_host` WHERE `type_id` = 3 AND `ip` IS NOT NULL ORDER BY `id` DESC")
	if err != nil {
		log.Fatal(err)
	}
	SipUserData := make([]SipUser, 0)
	err = dataBase.Select(&SipUserData, "SELECT `internalnumber`, `description`, `password` FROM `phones_phone` ORDER BY `id` DESC")
	if err != nil {
		log.Fatal(err)
	}
	tempString := ""
	for _, element := range PhoneData {
		tempString += "<option>" + element.Name + " " + element.Ip + " " + element.Mac + "</option>\n"
	}
	tempStringSip := ""
	for _, element := range SipUserData {
		tempStringSip += "<option>" + element.User + " " + element.Description + " " + element.Password + "</option>\n"
	}
	listPhone := Page{Phone: tempString, Title: "", SIP1: tempStringSip, SIP2: tempStringSip}
	w.Header().Set("Content-type", "text/html")
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, &listPhone)

}
func exec(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//fmt.Fprint(w, r.PostForm)
}

func main() {

	u1 := User{4999, "Ebay", "YAH5", true}
	u2 := User{0, "", "", false}
	x := [2]User{u1, u2}
	ph := Phone{"dlink", "000012121212", "122.123.123.132"}
	pC := PhoneConf{x, true, 52, true, 9, "2.0006"}
	_, err := pC.MakeConfig(&ph)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	err = loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/exec/", exec)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	err = http.ListenAndServe(":4004", nil)
	if err != nil {
		log.Fatal(err)
	}

}
