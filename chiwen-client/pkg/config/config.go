package config

import (
	"fmt"
	// fsnotify 配置文件监听
	"github.com/fsnotify/fsnotify"
	// viper 配置管理
	"github.com/spf13/viper"
)

func Init() (err error) {
	viper.SetConfigName("config") // 指定配置文件名称（不需要带后缀）
	viper.SetConfigType("yaml")   // 指定配置文件类型

	// 添加多个搜索路径，确保能从不同位置找到配置文件
	viper.AddConfigPath(".")             // 当前目录
	viper.AddConfigPath("./configs")     // 当前目录下的configs目录
	viper.AddConfigPath("../configs")    // 上一级目录下的configs目录
	viper.AddConfigPath("../../configs") // 上两级目录下的configs目录

	err = viper.ReadInConfig() // 读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return err
	}

	// 设置配置变更回调
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件修改了...")
	})
	return nil
}
