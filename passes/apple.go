package passes

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alvinbaena/passkit"
)

func GenerateCustomerPass() {
	c := passkit.NewStoreCard()
	c.AddAuxiliaryFields(passkit.Field{
		Key: "Header",
		Label: "Customer",
		Value:"1324KKAB",
	})
	// c.AddHeaderField(passkit.Field{
	// 	Key: "header_key",
	// 	Label: "header_label",
	// 	Value:"header_value",
	// })

	pass := passkit.Pass{
		FormatVersion:       1,
		TeamIdentifier:      "QC27AT556H",
		PassTypeIdentifier:  "pass.com.usual.customer",
		OrganizationName:    "Usual",
		SerialNumber:        "1234",
		Description:         "This is a test",
		StoreCard:         		c,
		Barcodes: []passkit.Barcode{
			{
				Format:          passkit.BarcodeFormatQR,
				Message:         "This is a test",
				MessageEncoding: "utf-8",
			},
		},
		BackgroundColor: "#ffffff",
		ForegroundColor: "#000000",
	}


	template := passkit.NewInMemoryPassTemplate()
	template.AddAllFiles("./passes/membership")

	signer := passkit.NewMemoryBasedSigner()
	signInfo, err := passkit.LoadSigningInformationFromFiles("./passes/private.p12", "never_know_your_next_move", "./passes/pass.cer")

	if err != nil {
		panic(err)
	}

	z, err := signer.CreateSignedAndZippedPassArchive(&pass, template, signInfo)
	if err != nil {
    	panic(err)
	}

	err = os.WriteFile("./passes/member.pkpass", z, 0644)
	if err != nil {
    	panic(err)
	}
}

func openImage(filename string) ([]byte) {
	fileToBeUploaded := filename
	file, err := os.Open(fileToBeUploaded)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	bytes := make([]byte, size)

	// read file into bytes
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)  
	if err != nil {
		panic(err)
	}
	return bytes
}