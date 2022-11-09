package main

import (
	"fmt"
	"os"

	"bits.chrsm.org/arpy"
)

func main() {
	_ = fmt.Println
	_ = os.Stdin

	fp, _ := os.OpenFile("test.rpa", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	rpa := arpy.New(0xDEADBEEF)
	rpa.AddFile("hello.txt", []byte("abcd, efg. hijk lmnop. qrs, tuv. wx, yz.\n"))
	rpa.WriteTo(fp)
	fp.Close()
}

func decode() {
	_ = fmt.Println
	_ = os.Stdin

	fp, err := os.Open("fonts.rpa")
	if err != nil {
		panic(err)
	}

	rpa, err := arpy.Decode(fp)

	content, err := rpa.FileAt(arpy.Index{
		Offset: 1055870,
		Size: 68548,
	})
	if err != nil { panic(err) }

	fp, err = os.OpenFile("hello.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil { panic(err) }
	fp.Write(content)
	fp.Close()
}
