package firconfig

import (
	"log"
	"os"

	"github.com/whosefriendA/firEtcd/pkg/firlog"

	yaml "gopkg.in/yaml.v3"
)

type Firconfig interface {
	Default()
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
