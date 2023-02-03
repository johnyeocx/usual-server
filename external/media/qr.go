package media

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/johnyeocx/usual/server/external/cloud"
)

var (
)

func GenerateSubscribeQRCode(s3Sess *session.Session, businessId int) {
	link := fmt.Sprintf(`https://usual.page.link/?link=https://usual.ltd/subscribe?business_id=%d
	&apn=com.usual.customer&afl=https://www.usual.ltd/subscribe?business_id=%d
	&isi=123456789&ibi=com.usual.customer&ifl=https://usual.ltd/subscribe?business_id=%d`, businessId, businessId, businessId)

	// open output file
	// qrCode, err := qrcode.New(link, qrcode.Low)
	qrCode, err := qr.Encode(link, qr.L, qr.Auto)

	if err != nil {
		log.Println(err)
		return
	}
	qr, _ := barcode.Scale(qrCode, 512, 512)
	
	// qr := qrCode.Image(256)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, qr)

	if err != nil {
		return
	}

	cloud.UploadImage(s3Sess, buf.Bytes(), "image/png", "/business/profile_qr/" + strconv.Itoa(businessId))
}