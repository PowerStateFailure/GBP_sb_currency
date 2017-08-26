package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/PowerStateFailure/GBP_sb_currency/sbdata"
)

func main() {
	const sberBaseAddr = "http://www.cbr.ru/scripts/XML_daily.asp?date_req="
	const datereqFormat = "02/01/2006"
	const roubleCharCode = "RUR"
	var valute string
	var valCurse sbdata.ValCurse
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
	err = sbdata.GetValCurse(sberBaseAddr+time.Now().Format(datereqFormat), &valCurse)

	if err != nil {
		fmt.Println("Failed to get data: ", err, "\nNothing to do")
		return
	}

	// TODO: move to func?
	if valute != roubleCharCode {
		valuteFound := false
		for i := range valCurse.Valute {
			if valCurse.Valute[i].CharCode == valute {
				valuteValue, _ = valCurse.Valute[i].GetValueNormalized()
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
			tmpCurrency, _ := valCurse.Valute[i].GetValueNormalized()
			fmt.Printf("%8.4f %s\n", float32(sourceValue)/(tmpCurrency/valuteValue), valCurse.Valute[i].CharCode)
		}
	}
}
