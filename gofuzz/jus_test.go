package test

import (
	"regexp"
	"testing"

	"github.com/xuri/excelize/v2"
)

func FuzzReverse(f *testing.F) {
	// testcases := []string{"a\n", "√", "ß∑∂ƒåß∂©˙", "˚©ƒ˙ƒ∂ˆ¥§®†¨∞´¶§"}
	// for _, tc := range testcases {
	// 	f.Add(tc) // Use f.Add to provide a seed corpus
	// }
	f.Fuzz(func(t *testing.T, orig string) {
		f := excelize.NewFile()
		re := regexp.MustCompile(`_x...._`)
		if re.MatchString(orig) {
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		index, err := f.NewSheet("Sheet2")
		if err != nil {
			panic(err)
		}
		f.SetCellValue("Sheet2", "A2", "Hello world.")
		f.SetCellValue("Sheet1", "B2", orig)
		f.SetActiveSheet(index)
		new, _ := f.GetCellValue("Sheet1", "B2")
		if new != orig {
			panic(err)
		}

	})
}
