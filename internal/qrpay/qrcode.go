package qrpay

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// DefaultQRSize is the default size of generated QR codes in pixels.
const DefaultQRSize = 200

// QRRecoveryLevel is the error recovery level for QR codes.
// Medium (15%) is recommended for printed media per SPAYD spec.
var QRRecoveryLevel = qrcode.Medium

// GenerateQRPNG generates a QR code as PNG bytes.
func GenerateQRPNG(content string, size int) ([]byte, error) {
	if size <= 0 {
		size = DefaultQRSize
	}

	png, err := qrcode.Encode(content, QRRecoveryLevel, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return png, nil
}

// GenerateQRBase64 generates a QR code and returns it as a Base64 data URL.
// The returned string can be used directly in an HTML img src attribute.
// Example: data:image/png;base64,iVBORw0KGgo...
func GenerateQRBase64(content string, size int) (string, error) {
	png, err := GenerateQRPNG(content, size)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + encoded, nil
}
