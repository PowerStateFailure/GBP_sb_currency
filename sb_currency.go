package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

// ValCurse - my super description
type ValCurse struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    time.Time
	Valute  []Valute
}

// Valute - my super description
type Valute struct {
	XMLName  xml.Name `xml:"Valute"`
	CharCode string
	// why string? 'cause comma-separated float  sucks in golang
	Value   string
	Nominal uint32
	Name    string
}

// Getting data from GET-response and returns it, its content type and error, if any
func getData(address string) ([]byte, string, error) {

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Get(address)

	if err != nil {
		return nil, "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, "", fmt.Errorf("Respone status not OK (%d)", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)

	return data, res.Header.Get("Content-Type"), err
}

func (v *Valute) getValueNormalized() (float32, error) {
	// TODO: redundant non-zero check?
	if v.Nominal == 0 {
		return 0.0, errors.New("Nominal is zero")
	}
	dotSeparatedFloat := strings.Replace(v.Value, ",", ".", 1)

	floatValue, err := strconv.ParseFloat(dotSeparatedFloat, 32)

	if err != nil {
		return 0.0, err
	}

	return float32(floatValue) / float32(v.Nominal), nil
}

func main() {
	const sberBaseAddr = "http://www.cbr.ru/scripts/XML_daily.asp?date_req="
	const datereqFormat = "02/01/2006"
	const roubleCharCode = "RUR"
	var sbData []byte
	var valute string
	var contentType string
	var valCurse ValCurse
	var err error
	var valuteValue float32
	var sourceValue int

	// init arg-parser
	flag.StringVar(&valute, "currency", "USD", "Валюта")
	flag.IntVar(&sourceValue, "value", 1, "Количество")
	flag.Parse()

	if sourceValue <= 0 {
		fmt.Println("Value must be positive")
		return
	}

	fmt.Println("Getting data...")
	sbData, contentType, err = getData(sberBaseAddr + time.Now().Format(datereqFormat))

	if err != nil {
		fmt.Println("Failed to get data: ", err, "\nNothing to do")
		return
	}

	fmt.Println("Parsing XML...")

	fmt.Println("Content-type is", contentType)

	// https://stackoverflow.com/questions/6002619/unmarshal-an-iso-8859-1-xml-input-in-go
	// decode xml with proper charset

	reader, err := charset.NewReader(bytes.NewReader(sbData), contentType)
	if err != nil {
		fmt.Println("Failed to init reader: ", err)
		return
	}

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&valCurse)

	if err != nil {
		fmt.Println("Failed to parse XML: ", err)
		return
	}

	// TODO: move to func?
	if valute != roubleCharCode {
		valuteFound := false
		for i := range valCurse.Valute {
			if valCurse.Valute[i].CharCode == valute {
				valuteValue, _ = valCurse.Valute[i].getValueNormalized()
				valuteFound = true
				break
			}
		}

		if !valuteFound {
			fmt.Println("Unknown currency")
			return
		}

		fmt.Printf("You've got %d %s (%.4f in %s)\n", sourceValue, valute, valuteValue*float32(sourceValue), roubleCharCode)
	} else {
		valuteValue = 1.0
		fmt.Printf("You've got %d %s\n", sourceValue, valute)
	}

	// 2nd step
	fmt.Println("In other currencies:")
	for i := range valCurse.Valute {
		if valCurse.Valute[i].CharCode != valute {
			tmpCurrency, _ := valCurse.Valute[i].getValueNormalized()
			fmt.Printf("%8.4f %s\n", float32(sourceValue)/(tmpCurrency/valuteValue), valCurse.Valute[i].CharCode)
		}
	}
}
