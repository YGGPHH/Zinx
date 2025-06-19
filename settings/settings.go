package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type ZinxConfig struct {
	Name             string `mapstructure:"name"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	Version          string `mapstructure:"version"`
	MaxPacketSize    uint32 `mapstructure:"max_packet_size"`
	MaxConn          int    `mapstructure:"max_conn"`
	WorkerPoolSize   uint32 `mapstructure:"worker_pool_size"`
	MaxWorkerTaskLen uint32 `mapstructure:"max_worker_task_len"`
	MaxMsgChanLen    uint32 `mapstructure:"max_msg_chan_len"`
}

var Conf = new(ZinxConfig)

func Init() (err error) {

	viper.SetConfigFile("../conf/config.yaml")

	viper.AddConfigPath(".")   // 指定查找配置文件的路径（这里使用相对路径）
	err = viper.ReadInConfig() // 读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return
	}
	// 把读取到的配置信息反序列化到 Conf 变量中
	if err = viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了...")
		if err = viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		}
	})
	return
}
