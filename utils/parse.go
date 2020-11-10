package utils

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/utils"
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
