package utils

import (
	"crypto/x509"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/context"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	Byte = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
	PiByte
	EiByte
)

const (
	IByte = 1
	KByte = IByte * 1000
	MByte = KByte * 1000
	GByte = MByte * 1000
	TByte = GByte * 1000
	PByte = TByte * 1000
	EByte = PByte * 1000
)

const (
	ReadBuff = 100 * KiByte
)

const (
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"
)

func ExtractNodeInfo(cert utils.Certificate) error {
	tlsConfig, err := utils.NewTLSConfigClient(cert)
	if err != nil {
		return err
	}
	if len(tlsConfig.Certificates) == 1 && len(tlsConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
		if err == nil {
			res := strings.SplitN(cert.Subject.CommonName, ".", 2)
			if len(res) != 2 || res[0] == "" || res[1] == "" {
				return fmt.Errorf("failed to parse node name from cert")
			} else {
				os.Setenv(context.KeyNodeName, res[1])
				os.Setenv(EnvKeyNodeNamespace, res[0])
			}
		} else {
			return fmt.Errorf("certificate format error")
		}
	}
	return nil
}

func Bytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(s, 1000, sizes)
}

func IBytes(s uint64) string {
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return humanateBytes(s, 1024, sizes)
}

func MakeKey(kind specv1.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "-" + name + "-" + ver
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(math.Log(float64(s)) / math.Log(base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

// CalculateModelTransferSize calculates estimated payload size in bytes for model parameters or gradients.
// Supports FP32 (32-bit), FP16 (16-bit), INT8 (8-bit), and block-scaled INT4 (4-bit with 32-bit scale per block).
func CalculateModelTransferSize(numParameters int64, nbits int, blockSize int) (int64, error) {
	if numParameters < 0 {
		return 0, fmt.Errorf("numParameters must be non-negative")
	}
	if blockSize <= 0 {
		blockSize = 64
	}
	switch nbits {
	case 32:
		return numParameters * 4, nil
	case 16:
		return numParameters * 2, nil
	case 8:
		numBlocks := (numParameters + int64(blockSize) - 1) / int64(blockSize)
		return numParameters + (numBlocks * 4), nil
	case 4:
		numBlocks := (numParameters + int64(blockSize) - 1) / int64(blockSize)
		packedBytes := (numParameters + 1) / 2
		return packedBytes + (numBlocks * 4), nil
	default:
		return 0, fmt.Errorf("unsupported bit precision: %d (choose 4, 8, 16, or 32)", nbits)
	}
}
