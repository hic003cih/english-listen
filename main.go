package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/domainr/whois"
	uuid "github.com/iris-contrib/go.uuid"

	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

const (
	signType string = "v3"
)

type DictWeb struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

type DictBasic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}
type DictResp struct {
	ErrorCode    string                 `json:"errorCode"`
	Query        string                 `json:"query"`
	Translation  []string               `json:"translation"`
	Basic        DictBasic              `json:"basic"`
	Web          []DictWeb              `json:"web,omitempty"`
	Lang         string                 `json:"l"`
	Dict         map[string]interface{} `json:"dict,omitempty"`
	Webdict      map[string]interface{} `json:"webdict,omitempty"`
	TSpeakUrl    string                 `json:"tSpeakUrl,omitempty"`
	SpeakUrl     string                 `json:"speakUrl,omitempty"`
	ReturnPhrase []string               `json:"returnPhrase,omitempty"`
}
type Response struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
type Config struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

var config Config
var fromLan string
var toLan string
var m = make(map[string]string)

func main() {

	http.HandleFunc("/", homeHandler)
	/* 	http.HandleFunc(
	        "/",
	        func(w http.ResponseWriter, r *http.Request) {
	            fileServer.ServeHTTP(w, r)
	        },
		) */
	http.HandleFunc("/whois", whoisHandler)
	http.HandleFunc("/sntransfer", sntransferHandler)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/check", checkHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fileServer := http.FileServer(http.Dir("static/"))
	fileServer.ServeHTTP(w, r)
}
func whoisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := r.PostFormValue("data")

	result, err := whoisQuery(data)

	if err != nil {
		jsonResponse(w, Response{Error: err.Error()})
		return
	}
	jsonResponse(w, Response{Result: result})
}
func sntransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := r.PostFormValue("data")

	result, err := sntransfer(data)

	if err != nil {
		jsonResponse(w, Response{Error: err.Error()})
		return
	}
	jsonResponse(w, Response{Result: result})
}
func sntransfer(data string) (string, error) {
	var tempString string
	//fmt.Println(data)
	//把文本中的\t \n \v \f \r 都分割成slice
	temp := strings.Fields(data)

	fmt.Println(temp)
	//s := make([]string, len(temp))
	for i := range temp[0 : len(temp)-1] {
		//相除2不等於0,表示不是2的倍數,是2的倍數的話是中文
		if ((i + 1) % 2) == 0 {
			continue
		} else {
			m[temp[i]] = temp[i+1]
		}

		//s[i] = fmt.Sprintf("%s"+temp[i]+"%s", "'", "',")
		//m["english"+string(i)] = temp[i]
		//fmt.Println(s[i])
		//fmt.Printf("map key is: %v ;value is:%v \n", m, m[temp[i]])
	}
	for k := range m {
		//chinese := ""
		//fmt.Println(k)
		//fmt.Scanln(&chinese)

		// Call the value by the key.

		/* for !(m[k] == chinese) {
			log.Print("Wrong value")
			fmt.Scanln(&chinese)
		} */

		//回傳每個單字的題目和輸入框給前端
		tempString = tempString + `<form id="answerForm">`
		tempString = tempString + fmt.Sprintf(k+"%s", `&emsp;<input type="text" id="input`+k+`" placeholder="english" />&emsp;<button type="submit">Check</button><br>`)
		tempString = tempString + `</form>`
	}
	//fmt.Println(tempString)
	return tempString, nil
}

func searchEnglish(wordContext string) {
	var curpath string = GetCurrentDirectory()
	err := InitConfig(curpath+"/script/config.json", &config)
	if err != nil {
		fmt.Println(err)
		fmt.Println("config.json is open error.")
		return
	}
	httpPost(wordContext, fromLan, toLan)
}
func whoisQuery(data string) (string, error) {
	fmt.Println(data)
	response, err := whois.Fetch(data)
	if err != nil {
		return "", err
	}
	return string(response.Body), nil
}
func jsonResponse(w http.ResponseWriter, x interface{}) {
	bytes, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	/* if r.Method == "GET" {
		time := time.Now().Unix()
		h := md5.New()
		h.Write([]byte(strconv.FormatInt(time, 10)))
		token := hex.EncodeToString(h.Sum(nil))
		t, _ := template.ParseFiles("./view/upload.ctpl")
		t.Execute(w, token)
	} else if */
	if r.Method == "POST" {
		//把上传的文件存储在内存和临时文件中
		r.ParseMultipartForm(32 << 20)
		//获取文件句柄，然后对文件进行存储等处理
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println("form file err: ", err)
			return
		}
		defer file.Close()
		//fmt.Fprintf(w, "%v", handler.Header)
		//创建上传的目的文件
		f, err := os.OpenFile("./files/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("open file err: ", err)
			return
		}
		defer f.Close()

		/* stat, err := f.Stat()
		if err != nil {
			panic(err)
		}

		var size = stat.Size()
		fmt.Println("file size=", size)

		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			fmt.Println(line)
			if err != nil {
				if err == io.EOF {
					fmt.Println("File read ok!")
					break
				} else {
					fmt.Println("Read file error!", err)
					return
				}
			}
		} */
		//拷贝文件
		io.Copy(f, file)

		file2, err := os.OpenFile("./files/"+handler.Filename, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println("Open file error!", err)
			return
		}
		defer file2.Close()

		stat, err := file2.Stat()
		if err != nil {
			panic(err)
		}

		var size = stat.Size()
		fmt.Println("file size=", size)

		buf := bufio.NewReader(file2)
		line2, err := buf.ReadString(' ')
		//fmt.Println(line2)
		fmt.Println(sntransfer(line2))
		/* for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			//fmt.Println(line)
			if err != nil {
				if err == io.EOF {
					fmt.Println("File read ok!")
					break
				} else {
					fmt.Println("Read file error!", err)
					return
				}
			}
		} */
	}
}
func httpPost(words, from, to string) {
	var err error
	u1, err2 := uuid.NewV4()
	if err2 != nil {
		fmt.Println(err2)
	}
	input := truncate(words)
	stamp := time.Now().Unix()
	instr := config.AppKey + input + u1.String() + strconv.FormatInt(stamp, 10) + config.AppSecret
	sig := sha256.Sum256([]byte(instr))
	var sigstr string = HexBuffToString(sig[:])

	data := make(url.Values, 0)
	data["q"] = []string{words}
	data["from"] = []string{from}
	data["to"] = []string{to}
	data["appKey"] = []string{config.AppKey}
	data["salt"] = []string{u1.String()}
	data["sign"] = []string{sigstr}
	data["signType"] = []string{signType}
	data["curtime"] = []string{strconv.FormatInt(stamp, 10)}

	var resp *http.Response
	resp, err = http.PostForm("https://openapi.youdao.com/api",
		data)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	//fmt.Println(string(body))
	var jsonObj DictResp
	json.Unmarshal(body, &jsonObj)
	//fmt.Println(jsonObj)

	show(&jsonObj, os.Stdout)
}
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func truncate(q string) string {
	res := make([]byte, 10)
	qlen := len([]rune(q))
	if qlen <= 20 {
		return q
	} else {
		temp := []byte(q)
		copy(res, temp[:10])
		lenstr := strconv.Itoa(qlen)
		res = append(res, lenstr...)
		res = append(res, temp[qlen-10:qlen]...)
		return string(res)
	}
}
func InitConfig(str string, cfg *Config) error {
	fileobj, err := os.Open(str)
	if err != nil {
		return err
	}

	defer fileobj.Close()

	var fileContext []byte
	fileContext, err = ioutil.ReadAll(fileobj)

	json.Unmarshal(fileContext, cfg)
	return nil
}
func show(resp *DictResp, w io.Writer) {
	if resp.ErrorCode != "0" {
		fmt.Fprintln(w, resp.ErrorCode)
		fmt.Fprintln(w, "请输入正确的数据")
	}
	fmt.Fprintln(w, "@", resp.Query)

	if resp.Basic.UkPhonetic != "" {
		fmt.Fprintln(w, "英:", "[", resp.Basic.UkPhonetic, "]")
	}
	if resp.Basic.UsPhonetic != "" {
		fmt.Fprintln(w, "美:", "[", resp.Basic.UsPhonetic, "]")
	}

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.Translation {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}
	fmt.Fprintln(w, "[延伸]")
	for key, item := range resp.Basic.Explains {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}

	fmt.Fprintln(w, "[网络]")
	for key, item := range resp.Web {
		fmt.Fprintln(w, "\t", key+1, ".", item.Key)
		fmt.Fprint(w, "\t翻译:")
		for _, val := range item.Value {
			fmt.Fprint(w, val, ",")
		}
		fmt.Fprint(w, "\n")
	}
}
func HexBuffToString(buff []byte) string {
	var ret string
	for _, value := range buff {
		str := strconv.FormatUint(uint64(value), 16)
		if len([]rune(str)) == 1 {
			ret = ret + "0" + str
		} else {
			ret = ret + str
		}
	}
	return ret
}
func checkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := r.PostFormValue("data")

	result, err := check(data)

	if err != nil {
		jsonResponse(w, Response{Error: err.Error()})
		return
	}
	jsonResponse(w, Response{Result: result})
}
func check(data string) (string, error) {

	if m["seal"] == data {
		return "答對", nil
	}

	return "錯誤", nil
}
