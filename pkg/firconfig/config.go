package firconfig

import (
	"log"
	"os"
	"reflect"

	"github.com/whosefriendA/firEtcd/pkg/firlog"

	yaml "gopkg.in/yaml.v3"
)

type Firconfig interface {
	Default()
	Validate() error
}

// TODO etcd dynamic config

func WriteRemote(conf Firconfig) error {

	return nil
}

// TODO etcd dynamic config
func ReadRemote(conf Firconfig) {

}

func Init(Path string, conf Firconfig) {
	_, err := os.Stat(Path)
	if err != nil {
		if os.IsNotExist(err) {
			conf.Default()
			WriteLocal(Path, conf)
			firlog.Logger.Warnf("please check for the %s if needed to be modified, then run again\n", Path)
		} else {
			firlog.Logger.Fatalln("config wrong err:", err)
		}
	}
	ReadLocal(Path, conf)

	// 应用环境变量覆盖
	applyEnvOverrides(conf)

	// 验证配置
	if err := conf.Validate(); err != nil {
		firlog.Logger.Fatalln("config validation failed:", err)
	}
}

func WriteLocal(Path string, conf Firconfig) error {
	out, err := yaml.Marshal(conf)
	if err != nil {
		firlog.Logger.Fatalln("failed to marshal config", Path, ":", err)
		return err
	}

	err = os.WriteFile(Path, out, 0644)
	if err != nil {
		firlog.Logger.Fatalln("failed to write ", Path, err)
		return err
	}
	return nil
}

func ReadLocal(Path string, conf Firconfig) error {
	log.Println("read from ", Path)
	data, err := os.ReadFile(Path)
	if err != nil {
		firlog.Logger.Fatalln("config.yml does not exist")
	}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		firlog.Logger.Fatalln("can't not read config.yml")
	}
	return err
}

// applyEnvOverrides 应用环境变量覆盖配置
func applyEnvOverrides(conf interface{}) {
	val := reflect.ValueOf(conf)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	applyEnvOverridesRecursive(val)
}

// applyEnvOverridesRecursive 递归应用环境变量覆盖
func applyEnvOverridesRecursive(val reflect.Value) {
	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 获取环境变量标签
		envTag := fieldType.Tag.Get("env")
		if envTag != "" && field.CanSet() {
			if envValue := os.Getenv(envTag); envValue != "" {
				switch field.Kind() {
				case reflect.String:
					field.SetString(envValue)
				case reflect.Bool:
					if envValue == "true" || envValue == "1" {
						field.SetBool(true)
					} else if envValue == "false" || envValue == "0" {
						field.SetBool(false)
					}
				}
			}
		}

		// 递归处理嵌套结构
		if field.Kind() == reflect.Struct {
			applyEnvOverridesRecursive(field)
		} else if field.Kind() == reflect.Ptr && !field.IsNil() {
			if field.Elem().Kind() == reflect.Struct {
				applyEnvOverridesRecursive(field.Elem())
			}
		}
	}
}
