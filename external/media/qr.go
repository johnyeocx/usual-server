package media

import (
	"log"

	"github.com/skip2/go-qrcode"
)

var (
	subscribeLink = `https://usual.page.link/?link=https://usual.ltd/subscribe?id=1&apn=com.usual.customer&afl=https://www.usual.ltd/subscribe?id=1&isi=123456789&ibi=com.usual.customer&ifl=https://usual.ltd/subscribe?id=1`
)

func GenerateSubscribeQRCode() {
	// var png []byte
	// png, err := qrcode.Encode("https://example.org", qrcode.Medium, 256)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// open output file
	err := qrcode.WriteFile(subscribeLink, qrcode.Medium, 256, "qr.png")
	if err != nil {
		log.Println(err)
		return
	}

}