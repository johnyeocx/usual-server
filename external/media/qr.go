package media

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/skip2/go-qrcode"
)

var (
)

func GenerateSubscribeQRCode(s3Sess *session.Session, businessId int) {
	link := fmt.Sprintf(`https://usual.page.link/?link=https://usual.ltd/subscribe?business_id=%d
	&apn=com.usual.customer&afl=https://www.usual.ltd/subscribe?business_id=%d
	&isi=123456789&ibi=com.usual.customer&ifl=https://usual.ltd/subscribe?business_id=%d`, businessId, businessId, businessId)

	// open output file
	qrCode, err := qrcode.New(link, qrcode.Medium)
	if err != nil {
		log.Println(err)
		return
	}
	
	image := qrCode.Image(512)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, image)

	if err != nil {
		return
	}

	cloud.UploadImage(s3Sess, buf.Bytes(), "image/png", "/business/profile_qr/" + strconv.Itoa(businessId))
}