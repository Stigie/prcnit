package main

import (
	"net/http"
	"html/template"
	textTemplate "text/template"
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
	Phone  []Phone
	SIP1   []SipUser
	SIP2   []SipUser
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

type server struct {
	Db *sqlx.DB
}

func (c *PhoneConf) MakeConfig(pf *Phone) (string, error) {
	latexTemplate, err := textTemplate.ParseFiles("TelConfig.xml")
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

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	PhoneData := make([]Phone, 0)
	err := s.Db.Select(&PhoneData, "SELECT `mac`, `name`, INET_NTOA(`ip`) AS ip FROM `unetmap_host` WHERE `type_id` = 3 AND `ip` IS NOT NULL ORDER BY `id` DESC")
	if err != nil {
		log.Fatal(err)
	}
	SipUserData := make([]SipUser, 0)
	err = s.Db.Select(&SipUserData, "SELECT `internalnumber`, `description`, `password` FROM `phones_phone` ORDER BY `id` DESC")
	if err != nil {
		log.Fatal(err)
	}
	listPhone := Page{Phone: PhoneData, Title: "", SIP1: SipUserData, SIP2: SipUserData}
	t, err := template.ParseFiles("index.html")
	t.Execute(w, &listPhone)

}
func (s *server) execHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//fmt.Fprint(w, r.PostForm)
}

func main() {
	flag.Parse()
	err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	s := server{
		Db: sqlx.MustConnect("mysql", config.MysqlLogin+":"+config.MysqlPassword+"@tcp("+config.MysqlHost+")/"+config.MysqlDb+"?charset=utf8"),
	}

	u1 := User{4999, "Ebay", "YAH5", true}
	u2 := User{0, "", "", false}
	x := [2]User{u1, u2}
	ph := Phone{"dlink", "000012121212", "122.123.123.132"}
	pC := PhoneConf{x, true, 52, true, 9, "2.0006"}
	_, err = pC.MakeConfig(&ph)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/exec/", s.execHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Print("Server started at port 4004")
	err = http.ListenAndServe(":4004", nil)
	if err != nil {
		log.Fatal(err)
	}

}
