package main

import (
	"bytes"
	"encoding/csv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

import "fmt"
import "net/http"
import "encoding/json"
import "time"
import "io/ioutil"

type Bitcoin struct {
	KRW float32
	USD float32
}

type SlackMsg struct {
	Text string `json:"text"`
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
		panic(err)
	}
}

func getBitcoinData() (Bitcoin, error) {
	var ret Bitcoin
	url := os.Getenv("BITCOIN_URL")
	resp, err := http.Get(url)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(buf, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func sendSlackMsg(msg string) {
	url := os.Getenv("SLACK_URL")
	slackMsg := SlackMsg{msg}
	jsonBytes, err := json.Marshal(slackMsg)
	reqBody := bytes.NewBufferString(string(jsonBytes))
	slackRes, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		panic(err)
	}
	defer slackRes.Body.Close()
}

func setCsvHeader(writer *csv.Writer) {
	data := []string{"Data", "Price"}
	err := writer.Write(data)

	checkError("Cannot write CSV Header.",err)
	defer writer.Flush()
}

func writeDataIntoCsv(writer *csv.Writer, price string, date string) {
	data := []string{date, price}
	err := writer.Write(data)
	checkError("Cannot write to file", err)

	defer writer.Flush()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(time.Minute)
	file, err := os.Create("result.csv")
	checkError("Cannot create file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	setCsvHeader(writer)

	for _ = range ticker.C {
		bitcoin, err := getBitcoinData()
		if err != nil {
			panic(err)
		}
		date := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("KRW: %f / USD: %f", bitcoin.KRW, bitcoin.USD)
		go sendSlackMsg(msg)
		go writeDataIntoCsv(writer, msg, date)
		fmt.Println(date + ": " + msg)
	}
	
}