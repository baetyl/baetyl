// Package fingerprint 指纹获取命令
package fingerprint

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

// BaetylAppID 机器指纹id 如需修改 对应baetyl-gateway下mian.go 和云端baetyl-cloud同步修改
const BaetylAppID = "tBJzpHWeCWkvUlGcBIqLWVFjuVrtiMVc"

func GetFingerprint() {
	info, err := utils.GetFingerprint(BaetylAppID)
	if err != nil {
		log.L().Error("get fingerprint err: " + err.Error())
		return
	}
	fmt.Println("Copy The Fingerprint:")
	fmt.Println(info)
}
