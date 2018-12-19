package main

import (
	"fmt"
	"log"
	"os"
	"bufio"
	"encoding/json"
	"sync"
	"strings"
	"flag"
	"io/ioutil"
)

type answerType struct {
	Ttl int32   `json:"ttl"`
	AnswerType string   `json:"type"`
	Class string    `json:"class"`
	Name string     `json:"name"`
	Answer string   `json:"answer"`
}

type flagType struct {
	Response bool   `json:"response"`
	Opcode int32    `json:"opcode"`
	Authoritative bool  `json:"authoritative"`
	Truncated bool  `json:"truncated"`
	Recursion_desired bool  `json:"recursion_desired"`
	Recursion_available bool    `json:"recursion_available"`
	Authenticated bool  `json:"authenticated"`
	Checking_disabled bool  `json:"checking_disabled"`
	Error_code int32    `json:"error_code"`
}

type dataType struct {
	Answers [] answerType   `json:"answers"`
	Additionals []string    `json:"additionals"`
	Authorities []string    `json:"authorities"`
	Protocol string         `json:"protocol"`
	Flags flagType          `json:"flags"`
}
type record struct {
	Name string         `json:"name"`
	Class string        `json:"class"`
	Status string       `json:"status"`
	Timestamp string    `json:"timestamp"`
	Data dataType       `json:"data"`
}

var (
	dirPath = flag.String("dirPath",".","Path of the input directory.")
)

func parseJsonString (fileName string, plainStrings chan* string, answerRecords chan* string, wg *sync.WaitGroup) {
	var result record
	for line := range plainStrings {
		json.Unmarshal([]byte(*line), &result)
		if result.Status == "NXDOMAIN" {
			answerRecord := "NXDOMAIN" + " " + result.Name
			answerRecords <- &answerRecord
		}
		for _, answer:= range result.Data.Answers {
			if len(answer.Name) == 0 || len(answer.Answer) == 0 {
				continue
			}
			answerRecord := answer.AnswerType + " " + answer.Name + "," + answer.Answer
			answerRecords <- &answerRecord
			//fmt.Println(answer.AnswerType, answer.Name, answer.Answer)
		}
	}
	close(answerRecords)
	(*wg).Done()
}

func writeToFile (fileName string, answerRecords chan* string, wg *sync.WaitGroup) {
	for answerRecord := range(answerRecords) {
		parts := strings.Split(*answerRecord, " ")
		f, err := os.OpenFile(fileName[:len(fileName)-5]+"_"+parts[0]+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.WriteString(parts[1] + "\n")
		f.Close()
		//fmt.Println(*answerRecord)
	}
	(*wg).Done()
}

func parseJsonFile(fileName string) {
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()
	// build parse json routine
	plainStrings := make(chan *string)
	answerRecords := make(chan *string)
	var WG sync.WaitGroup
	WG.Add(2)
	go parseJsonString(fileName, plainStrings, answerRecords, &WG)
	go writeToFile(fileName, answerRecords,  &WG)

	scanner := bufio.NewScanner(jsonFile)
	for scanner.Scan() {
		line := scanner.Text()
		plainStrings <- &line
	}
	close(plainStrings)
	WG.Wait()
	fmt.Println("Complete reading file " + fileName)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	files, err := ioutil.ReadDir(*dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fName := file.Name()
		if len(fName) > 5 && fName[len(fName)-5:] == ".json" {
			parseJsonFile(fName)
		}
	}
}