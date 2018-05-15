package utils

import "github.com/astaxie/beego/logs"

var ConsoleLogs *logs.BeeLogger
var FileLogs *logs.BeeLogger

func init() {
	ConsoleLogs = logs.NewLogger(1000)
	ConsoleLogs.SetLogger("console")
	ConsoleLogs.SetLevel(logs.LevelDebug) // 设置日志写入缓冲区的等级：Debug级别（最低级别，所以所有log都会输入到缓冲区）
	ConsoleLogs.EnableFuncCallDepth(true) // 输出log时能显示输出文件名和行号（非必须）
	FileLogs = logs.NewLogger(1000)
	FileLogs.SetLogger("file", `{"filename":"logs/httpRequestLog.log"}`)
	FileLogs.SetLevel(logs.LevelDebug) // 设置日志写入缓冲区的等级：Debug级别（最低级别，所以所有log都会输入到缓冲区）
	FileLogs.EnableFuncCallDepth(true) // 输出log时能显示输出文件名和行号（非必须）
}
