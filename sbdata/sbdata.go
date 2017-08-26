package sbdata

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"

	"github.com/PowerStateFailure/GBP_sb_currency/netget"
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

func (v *Valute) GetValueNormalized() (float32, error) {
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

func GetValCurse(address string, vc *ValCurse) error {
	rawData, contentType, err := netget.GetData(address)

	if err != nil {
		return err
	}

	// https://stackoverflow.com/questions/6002619/unmarshal-an-iso-8859-1-xml-input-in-go
	// decode xml with proper charset

	reader, err := charset.NewReader(bytes.NewReader(rawData), contentType)
	if err != nil {
		fmt.Println("Failed to init reader: ", err)
		return err
	}

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(vc)

	return err
}
