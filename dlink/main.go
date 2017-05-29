package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	sw "strings"

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
	EnableReg bool
}

type PhoneConf struct {
	Users         [2]User
	VlanPhone     bool
	VlanPhoneNumb int
	VlanComp      bool
	VlanCompNumb  int
	Version       float64
}

type server struct {
	Db *sqlx.DB
}

//function from makeConfig pakage
func newline() string {
	return "\n            "
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (c *PhoneConf) MakeConfig(pf *Phone) (string, error) {
	fmap := template.FuncMap{
		"addOne":  addOne,
		"newline": newline,
	}
	latexTemplate := template.Must(template.New("Config template").Funcs(fmap).ParseFiles("TelConfig.xml"))
	//latexTemplate, err := template
	/*if err != nil {
		return "", err
	}*/

	if path, err := exists("userFiles/" + pf.Mac + ".xml"); err == nil && path == true {
		file, _ := os.Open("userFiles/" + pf.Mac + ".xml")
		f := bufio.NewReader(file)
		var i = 0;
		var read_line string;
		for i < 3 {
			read_line, _ = f.ReadString('\n')
			i++;
		}

		file.Close()
		start := sw.Index(read_line, ">");
		read_line = read_line[start+1:len(read_line)];
		start = sw.Index(read_line, "<");
		read_line = cut(read_line, start);
		version, err := strconv.ParseFloat(read_line, 64)
		if err != nil {
			log.Println(err)
		}
		version+=0.0001
		//fmt.Println(version);
		c.Version = version;

	}

	outputLatexFile, err := os.Create("userFiles/" + pf.Mac + ".xml")
	if err != nil {
		return "", err
	}
	//log.Print(err);
	defer outputLatexFile.Close()
	err = latexTemplate.ExecuteTemplate(outputLatexFile, "TelConfig.xml", createStatement(c))
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles\\" + pf.Mac + ".xml"
	return pathToTexFile, nil
}

func addOne(number int) int {
	return number + 1
}

func createStatement(c *PhoneConf) PhoneConf {
	return PhoneConf{
		VlanPhone:     c.VlanPhone,
		VlanPhoneNumb: c.VlanPhoneNumb,
		VlanComp:      c.VlanComp,
		VlanCompNumb:  c.VlanCompNumb,
		Version:       c.Version,
		Users: [2]User{
			User{
				UserLog:   c.Users[0].UserLog,
				UserPas:   c.Users[0].UserPas,
				UserName:  c.Users[0].UserName,
				EnableReg: c.Users[0].EnableReg,

			},
			User{
				UserLog:   c.Users[1].UserLog,
				UserPas:   c.Users[1].UserPas,
				UserName:  c.Users[1].UserName,
				EnableReg: c.Users[1].EnableReg,

			},
		},
	}
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
		log.Print(err)
		return
	}
	SipUserData := make([]SipUser, 0)
	err = s.Db.Select(&SipUserData, "SELECT `internalnumber`, `description`, `password` FROM `phones_phone` ORDER BY `id` DESC")
	if err != nil {
		log.Print(err)
		return
	}
	listPhone := Page{Phone: PhoneData, Title: "", SIP1: SipUserData, SIP2: SipUserData}
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Print(err)
		return
	}
	t.Execute(w, &listPhone)

}
func (s *server) execHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	fmt.Fprint(w, r.PostForm)
	if err != nil {
		log.Print(err)
		return
	}
	_ = r.PostFormValue("ff")
	forma := r.PostForm;
	var SIP1, SIP2, vlanIDtel, vlanIDcomp int;
	if SIP1, err = strconv.Atoi(forma.Get("userPhoneNumber1")); err != nil || (SIP1 < 1000 && SIP1 > 9999) {
		log.Println(err)
		return
	}
	if SIP2, err = strconv.Atoi(forma.Get("userPhoneNumber2")); err != nil || (SIP2 < 1000 && SIP2 > 9999) {
		log.Println(err)
		return
	}

	if vlanIDtel, err = strconv.Atoi(forma.Get("vlanIDtel")); err != nil || (vlanIDtel < 1 && vlanIDtel > 4095) {
		log.Println(err)
		return
	}
	if vlanIDcomp, err = strconv.Atoi(forma.Get("vlanIDcomp")); err != nil || (vlanIDcomp < 1 && vlanIDcomp > 4095) {
		log.Println(err)
		return
	}
	log.Print(SIP1);
	log.Print(SIP2);
	log.Print(vlanIDcomp);
	log.Println(vlanIDtel);
	MacAddr := forma.Get("phone");
	start := sw.Index(MacAddr, " ");
	MacAddr = MacAddr[start+1:len(MacAddr)];
	start = sw.Index(MacAddr, " ");
	MacAddr = cut(MacAddr, start);
	MacAddr = sw.ToUpper(MacAddr);
	log.Println(MacAddr)
	rows, err := s.Db.NamedQuery("SELECT `internalnumber`, `description`, `password` FROM `phones_phone` WHERE `internalnumber` = :sip1 OR  `internalnumber` = :sip2", map[string]interface{}{"sip1": SIP1, "sip2": SIP2})
	if err != nil {
		log.Print(err)
		return
	}
	phoneData := SipUser{}
	SipUserData := []SipUser{}
	for rows.Next() {
		err := rows.StructScan(&phoneData)
		if err != nil {
			log.Print(err)
			return
		}
		SipUserData = append(SipUserData, phoneData)

	}
	log.Println(SipUserData)
	//log.Println(phoneData[2])

	u1 := User{SIP1, SipUserData[0].Password, SipUserData[0].Description, true}
	u2 := User{}
	if (forma.Get("SecondSipEnable") == "on"){
		u2 = User{SIP2, SipUserData[1].Password, SipUserData[1].Description, true}
	}else {
		u2 = User{0, "", "", false}
	}
	x := [2]User{u1, u2}
	ph := Phone{forma.Get("UserPhone"), MacAddr, "122.123.123.132"}
	pC := PhoneConf{x, true, 52, true, 9, 2.0006}
	_, err = pC.MakeConfig(&ph)
	if err != nil {
		log.Fatal(err)
	}


	//log.Println(phoneData.);
	//log.Println(SipUserData[1].Password);
}
func cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
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
