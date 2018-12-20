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
	"time"
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
	inputDirPath = flag.String("inDirPath",".","Path of the input directory.")
	outputDirPath = flag.String("outDirPath",".","Path of the output directory.")
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
	fullPath := ""
	if len(*outputDirPath) > 0 && (*outputDirPath)[len(*outputDirPath)-1] != '/' {
		fullPath = (*outputDirPath) + "/" + fileName
	} else {
		fullPath = (*outputDirPath) + fileName
	}
	fpDict := make(map[string]*os.File)
	for answerRecord := range(answerRecords) {
		parts := strings.Split(*answerRecord, " ")
		f, ok := fpDict[parts[0]]
		if !ok {
			ftmp, err := os.OpenFile(fullPath[:len(fullPath)-5]+"_"+parts[0]+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			} else {
				defer ftmp.Close()
				fpDict[parts[0]] = ftmp
				f, _ = fpDict[parts[0]]
			}
		}

		f.WriteString(parts[1] + "\n")

		//fmt.Println(*answerRecord)
	}
	for _, f := range fpDict {
		f.Close()
	}
	(*wg).Done()
}

func parseJsonFile(fileName string) {
	fullPath := ""
	if len(*inputDirPath) > 0 && (*inputDirPath)[len(*inputDirPath)-1] != '/' {
		fullPath = (*inputDirPath) + "/" + fileName
	} else {
		fullPath = (*inputDirPath) + fileName
	}

	jsonFile, err := os.Open(fullPath)
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
	start := time.Now()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	files, err := ioutil.ReadDir(*inputDirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fName := file.Name()
		if len(fName) > 5 && fName[len(fName)-5:] == ".json" {
			parseJsonFile(fName)
		}
	}
	elapsed := time.Since(start)
	timeCostFile, err := os.Create("timeCost.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer timeCostFile.Close()
	w := bufio.NewWriter(timeCostFile)
	w.WriteString(elapsed.String())
	w.Flush()
	fmt.Println("Finished")
}