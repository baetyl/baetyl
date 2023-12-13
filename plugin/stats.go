// Package plugin 定义各类插件接口
package plugin

type Collect interface {
	CollectStats(mode string) (map[string]interface{}, error)
}
