package log

import (
	"testing"
	"encoding/json"
	"fmt"
	"time"
)

var log *Log


/*func TestMain(m *testing.M){
	config := make(map[string]interface{})
	config["filename"] = "./logcollect.log"
	//config["level"] = LevelDebug
	config["maxsize"] = 268436
	json,err:= json.Marshal(config)
	if err != nil {
		panic(err)
	}
	fmt.Println("testMain")
	SetLogger("file",string(json))
	SetLevel(LevelDebug)
	//log = Async(10000)
	m.Run()
}*/

/*
func TestLog(t *testing.T){
	t.Run("debug",testDebug)
	t.Run("info",testInfo)
	t.Run("notice",testNotice)
	t.Run("warnning",testWarnning)
	t.Run("error",testError)
	t.Run("critical",testCritical)
	t.Run("alert",testAlert)
	t.Run("emerygency",testEmergency)

	log.Close()
	time.Sleep(time.Second)
}
*/
func testDebug(t *testing.T){
	Debug("debug")
}

func testInfo(t *testing.T) {
	Info("info")
}

func testNotice(t *testing.T) {
	Notice("info")
}

func testWarnning(t *testing.T){
	Warning("warning")
}

func testError(t *testing.T){
	Error("warning")
}

func testCritical(t *testing.T) {
	Critical("critical")
}

func testAlert(t *testing.T){
	Alert("alert")
}

func testEmergency(t *testing.T) {
	Emergency("emeryency")
}


func TestInfo(t *testing.T){

	config := make(map[string]interface{})
	config["filename"] = "./logcollect.log"
	//config["level"] = LevelDebug
	config["maxsize"] = 268436
	json,err:= json.Marshal(config)
	if err != nil {
		panic(err)
	}
	fmt.Println("testMain")
	SetLogger("file",string(json))
	SetLevel(LevelDebug)

	for i:=0;i<10;i++{
		go func(i int){
			for j := 0;j<10000;j++ {
				Info("the values is %s","gary")
				Warning("warning1")
				Debug("debug")
				Info("info1")
				Info("info2")
				Warning("warning2")
				Warning("warning3")
				Alert("alert")
			}
			if i == 5 || i == 7 {
				time.Sleep(500 * time.Millisecond)
			}
		}(i)
	}


	time.Sleep(10*time.Second)
	//log.Close()

}


