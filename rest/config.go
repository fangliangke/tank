package rest

import (
	"encoding/json"
	"github.com/json-iterator/go"
	"io/ioutil"
	"time"
	"unsafe"
)

const (
	//用户身份的cookie字段名
	COOKIE_AUTH_KEY = "_ak"

	//数据库表前缀 tank200表示当前应用版本是tank:2.0.x版，数据库结构发生变化必然是中型升级
	TABLE_PREFIX = "tank20_"

	//当前版本
	VERSION = "2.0.0"
)

/*
如果你需要在本地127.0.0.1创建默认的数据库和账号，使用以下语句。
create database tank;
grant all privileges on tank.* to tank identified by 'tank123';
flush privileges;
*/
var CONFIG = &Config{}

//依赖外部定义的变量。
type Config struct {
	//默认监听端口号
	ServerPort int
	//数据库是否配置完备
	DBConfigured bool
	//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
	MatterPath string
	//数据库连接信息。
	MysqlUrl string
	//配置文件中的项
	Item *ConfigItem
}

//和tank.json文件中的键值一一对应。
type ConfigItem struct {
	//默认监听端口号
	ServerPort int
	//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
	MatterPath string
	//mysql相关配置。
	//数据库端口
	MysqlPort int
	//数据库Host
	MysqlHost string
	//数据库名字
	MysqlSchema string
	//用户名
	MysqlUsername string
	//密码
	MysqlPassword string
}

//验证配置文件的正确性。
func (this *ConfigItem) validate() bool {

	if this.ServerPort == 0 {
		LOGGER.Error("ServerPort 未配置")
		return false
	}

	if this.MysqlUsername == "" {
		LOGGER.Error("MysqlUsername 未配置")
		return false
	}

	if this.MysqlPassword == "" {
		LOGGER.Error("MysqlPassword 未配置")
		return false
	}

	if this.MysqlHost == "" {
		LOGGER.Error("MysqlHost 未配置")
		return false
	}

	if this.MysqlPort == 0 {
		LOGGER.Error("MysqlPort 未配置")
		return false
	}

	if this.MysqlSchema == "" {
		LOGGER.Error("MysqlSchema 未配置")
		return false
	}

	return true

}

//验证配置文件是否完好
func (this *Config) Init() {

	//JSON初始化
	jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		//如果使用time.UTC，那么时间会相差8小时
		t, err := time.ParseInLocation("2006-01-02 15:04:05", iter.ReadString(), time.Local)
		if err != nil {
			iter.Error = err
			return
		}
		*((*time.Time)(ptr)) = t
	})

	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		//如果使用time.UTC，那么时间会相差8小时
		stream.WriteString(t.Local().Format("2006-01-02 15:04:05"))
	}, nil)

	//默认从6010端口启动
	CONFIG.ServerPort = 6010

	//读取配置文件
	filePath := GetConfPath() + "/tank.json"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		LOGGER.Warn("即将进入安装过程，无法找到配置文件：%s", filePath)
		this.DBConfigured = false
	} else {
		// 用 json.Unmarshal
		err := json.Unmarshal(content, this.Item)
		if err != nil {
			LOGGER.Error("配置文件格式错误！")
			this.DBConfigured = false
			return
		}

		//验证项是否齐全
		itemValidate := this.Item.validate()
		if !itemValidate {
			this.DBConfigured = false
			return
		}

		if this.Item.MatterPath == "" {
			CONFIG.MatterPath = GetHomePath() + "/matter"
		}
		MakeDirAll(CONFIG.MatterPath)

		this.MysqlUrl = GetMysqlUrl(this.Item.MysqlPort, this.Item.MysqlHost, this.Item.MysqlSchema, this.Item.MysqlUsername, this.Item.MysqlPassword)
		this.DBConfigured = true
	}

}
