package version

// 编译时注入的版本变量
// 可以通过 ldflags 在编译时设置，例如：
// go build -ldflags "-X github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/version.Version=1.0.0"
var Version = "1.0.0"

// GetVersion 返回当前版本号
func GetVersion() string {
	return Version
}
