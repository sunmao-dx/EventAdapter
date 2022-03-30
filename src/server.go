package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// "os"

	gitee_utils "gitee.com/sunmao-dx/strategy-executor/src/gitee-utils"
	"github.com/sirupsen/logrus"
)

var eventTypeMap map[string]string

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Event received.")

	// // Loop over header names
	// for name, values := range r.Header {
	// 	// Loop over all values for the name.
	// 	for _, value := range values {
	// 		fmt.Println(name, value)
	// 	}
	// }
	// fmt.Println(r.Body)

	eventType, _, payload, ok, _ := gitee_utils.ValidateWebhook(w, r)
	if !ok {
		gitee_utils.LogInstance.WithFields(logrus.Fields{
			"context": "gitee hook is broken",
		}).Info("info log")
		return
	}

	strApi, ok := eventTypeMap[eventType]
	if ok {
		// strApi := os.Getenv("api_url")
		// strApi := "http://localhost:8001"

		_, err := sendRequest(r.Header, payload, strApi)

		if err != nil {
			gitee_utils.LogInstance.WithFields(logrus.Fields{
				"context": "Send " + eventType + " problem",
			}).Info("info log")
			fmt.Println(err.Error())
		}
		gitee_utils.LogInstance.WithFields(logrus.Fields{
			"context": "Send " + eventType + " success",
		}).Info("info log")
	} else {
		fmt.Println(eventType + " does not exist, please check again")
		return
	}

}

func makeMapfromJson(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file err = ", err)
		return
	}

	defer f.Close()

	// eventTypeMap := make(map[string]string)
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&eventTypeMap)
	if err != nil {
		fmt.Println("json decode has error: ", err)
	}
}

func sendRequest(header http.Header, payload []byte, apiUrl string) (string, error) {
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(payload))
	if err != nil {
		return "Bad Request", err
	}

	req.Header = header

	client := &http.Client{}

	var retries = 3
	for retries > 0 {
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			gitee_utils.LogInstance.WithFields(logrus.Fields{
				"context": "Send " + header.Get("X-Gitee-Event") + " error, please check whether the apiUrl is active",
			}).Info("info log")
			retries -= 1
		} else {
			defer resp.Body.Close()
			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("response Body:", string(body))
			break
		}
	}
	return "Good Request", nil
}

func main() {
	filePath := os.Getenv("filePath")
	// filePath := "src/config.json"
	makeMapfromJson(filePath)

	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8000", nil)
}

// $ echo "export GO111MODULE=on" >> ~/.bashrc
// $ echo "export GOPROXY=https://goproxy.cn" >> ~/.bashrc
// $ source ~/.bashrc
