package passes

import (
	"fmt"

	"github.com/alvinbaena/passkit"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/johnyeocx/usual/server/external/cloud"
)

func GenerateCustomerPass(
	s3sess *session.Session,
	cusName string, 
	cusUuid string,
	cusId int,
) (error) {
	c := passkit.NewStoreCard()
	c.AddAuxiliaryFields(passkit.Field{
		Key: "Header",
		Label: "Customer",
		Value: cusName,
	})

	pass := passkit.Pass{
		FormatVersion:       1,
		TeamIdentifier:      "QC27AT556H",
		PassTypeIdentifier:  "pass.com.usual.customer",
		OrganizationName:    "Usual",
		SerialNumber:        cusUuid,
		Description:         "Usual membership card",
		StoreCard:         	 c,
		Barcodes: []passkit.Barcode{
			{
				Format:          passkit.BarcodeFormatQR,
				Message:         cusUuid,
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
		return err
	}

	z, err := signer.CreateSignedAndZippedPassArchive(&pass, template, signInfo)
	if err != nil {
    	return err
	}
	
	key := fmt.Sprintf("customer/pkpass/%d.pkpass", cusId)
	err = cloud.PutObject(s3sess, z, "binary/octet-stream", key)
	return err
}

// func openImage(filename string) ([]byte) {
// 	fileToBeUploaded := filename
// 	file, err := os.Open(fileToBeUploaded)

// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	defer file.Close()

// 	fileInfo, _ := file.Stat()
// 	var size int64 = fileInfo.Size()
// 	bytes := make([]byte, size)

// 	// read file into bytes
// 	buffer := bufio.NewReader(file)
// 	_, err = buffer.Read(bytes)  
// 	if err != nil {
// 		panic(err)
// 	}
// 	return bytes
// }