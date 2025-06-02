package firconfig

import (
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"log"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type LaneConfig interface {
	Default()
}

// TODO etcd dynamic config

func WriteRemote(conf LaneConfig) error {

	return nil
}

// TODO etcd dynamic config
func ReadRemote(conf LaneConfig) {

}

func Init(Path string, conf LaneConfig) {
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

func WriteLocal(Path string, conf LaneConfig) error {
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

func ReadLocal(Path string, conf LaneConfig) error {
	log.Println("read from ", Path)
	data, err := os.ReadFile(Path)
	if err != nil {
		firlog.Logger.Fatalln("config.yaml does not exist")
	}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		firlog.Logger.Fatalln("can't not read config.yml")
	}
	return err
}
