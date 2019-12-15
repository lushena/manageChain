package logs

import (
	"encoding/json"
	"fmt"
	gglogs "gglogs"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Log levels to control the logging output.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

const defaultFormat = "%s"

type FabricLogger struct {
	logger *gglogs.BeeLogger
}

type ggLogInfo struct {
	isOpenGglog     bool
	filepath        string
	filename        string
	maxlinesPerFile int
	maxsizePerFile  int
	maxTotalSize    int64
	isAutoDelete    bool
	daily           bool
	rotate          bool
	maxdays         int
}

const (
	Orderer_Prefix = "ORDERER"
	Peer_Prefix    = "CORE"
	FilenameLenMax = 128
	FilepathLenMax = 128

	//===============DefaultArg====================
	DefaultCCFileName      = "cc_log"
	DefaultFilePath        = "/var/fabric_logs"
	DefaultFileName        = "gg_log"
	DefaultMaxlinesPerFile = 10000000
	DefaultMaxsizePerFile  = 102400000
	DefaultMaxTotalSize    = 4096000000
	DefaultLogLevel        = LevelInformational
	DefaultMaxDays         = 15
)

var once sync.Once
var fl *FabricLogger
var loglevel int = LevelDebug
var levelNames = [...]string{"emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"}

func SetFabricLogger(containerType string) *FabricLogger {
	once.Do(func() {
		var fabricCfgPath = os.Getenv("FABRIC_CFG_PATH")
		var configName string
		var loginfo ggLogInfo

		if containerType == "orderer" {
			configName = strings.ToLower(Orderer_Prefix)
			config := viper.New()
			config.SetConfigName(configName)
			config.AddConfigPath(fabricCfgPath)
			config.ReadInConfig()
			config.SetEnvPrefix("ORDERER")
			config.AutomaticEnv()
			replacer := strings.NewReplacer(".", "_")
			config.SetEnvKeyReplacer(replacer)
			config.SetConfigType("yaml")

			var err error
			loginfo.isOpenGglog, err = strconv.ParseBool(config.GetString("general.isOpenGglog"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.isOpenGglog to bool, err: %s", err))
			}

			loginfo.filepath = config.GetString("general.logpath")
			loginfo.filename = config.GetString("general.logname")

			loginfo.maxlinesPerFile, err = strconv.Atoi(config.GetString("general.maxlinesPerFile"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.maxlinesPerFile to int, err: %s", err))
			}
			loginfo.maxsizePerFile, err = strconv.Atoi(config.GetString("general.maxsizePerFile"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.maxsizePerFile to int, err: %s", err))
			}
			loginfo.maxTotalSize, err = strconv.ParseInt(config.GetString("general.maxTotalSize"), 10, 64)
			if err != nil {
				panic(fmt.Sprintf("Error convert general.maxTotalSize to int64, err: %s", err))
			}
			loginfo.maxdays, err = strconv.Atoi(config.GetString("general.maxdays"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.maxdays to int, err: %s", err))
			}
			loginfo.daily, err = strconv.ParseBool(config.GetString("general.daily"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.daily to bool, err: %s", err))
			}
			loginfo.rotate, err = strconv.ParseBool(config.GetString("general.rotate"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.rotate to bool, err: %s", err))
			}
			loginfo.isAutoDelete, err = strconv.ParseBool(config.GetString("general.isautodelete"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.isautodelete to bool, err: %s", err))
			}
			loglevel, err = strconv.Atoi(config.GetString("general.ggLogLevel"))
			if err != nil {
				panic(fmt.Sprintf("Error convert general.ggLogLevel to int, err: %s", err))
			}
			// fmt.Printf("orderer loglevel:%d\n\n", loglevel)
		} else if containerType == "peer" {
			configName = strings.ToLower(Peer_Prefix)
			config := viper.New()
			config.SetConfigName(configName)
			config.AddConfigPath(fabricCfgPath)
			config.ReadInConfig()
			config.SetEnvPrefix("CORE")
			config.AutomaticEnv()
			replacer := strings.NewReplacer(".", "_")
			config.SetEnvKeyReplacer(replacer)
			config.SetConfigType("yaml")

			var err error
			loginfo.isOpenGglog, err = strconv.ParseBool(config.GetString("logging.isOpenGglog"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.isOpenGglog to bool, err: %s", err))
			}

			loginfo.filepath = config.GetString("logging.logpath")
			loginfo.filename = config.GetString("logging.logname")

			loginfo.maxlinesPerFile, err = strconv.Atoi(config.GetString("logging.maxlinesPerFile"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.maxlinesPerFile to int, err: %s", err))
			}
			loginfo.maxsizePerFile, err = strconv.Atoi(config.GetString("logging.maxsizePerFile"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.maxsizePerFile to int, err: %s", err))
			}
			loginfo.maxTotalSize, err = strconv.ParseInt(config.GetString("logging.maxTotalSize"), 10, 64)
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.maxTotalSize to int64, err: %s", err))
			}
			loginfo.maxdays, err = strconv.Atoi(config.GetString("logging.maxdays"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.maxdays to int, err: %s", err))
			}
			loginfo.daily, err = strconv.ParseBool(config.GetString("logging.daily"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.daily to bool, err: %s", err))
			}
			loginfo.rotate, err = strconv.ParseBool(config.GetString("logging.rotate"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.rotate to bool, err: %s", err))
			}
			loginfo.isAutoDelete, err = strconv.ParseBool(config.GetString("logging.isautodelete"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.isautodelete to bool, err: %s", err))
			}
			loglevel, err = strconv.Atoi(config.GetString("logging.ggLogLevel"))
			if err != nil {
				panic(fmt.Sprintf("Error convert logging.ggLogLevel to int, err: %s", err))
			}
			// fmt.Printf("peer loglevel:%d\n\n", loglevel)
		} else if containerType == "chaincode" {
			var err error
			loglevel, err = strconv.Atoi(os.Getenv("CHAINCODE_LOG_LEVEL"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_LEVEL to int, err: %s", err))
			}
			loginfo.isOpenGglog, err = strconv.ParseBool(os.Getenv("CHAINCODE_LOG_ISOPENGGLOG"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_ISOPENGGLOG to bool, err: %s", err))
			}

			loginfo.filepath = os.Getenv("CHAINCODE_LOG_DESTINATION")
			loginfo.filename = DefaultCCFileName

			loginfo.maxlinesPerFile, err = strconv.Atoi(os.Getenv("CHAINCODE_LOG_MAXLINES"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_MAXLINES to int, err: %s", err))
			}
			loginfo.maxsizePerFile, err = strconv.Atoi(os.Getenv("CHAINCODE_LOG_MAXSIZE"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_MAXSIZE to int, err: %s", err))
			}
			loginfo.maxTotalSize, err = strconv.ParseInt(os.Getenv("CHAINCODE_LOG_MAXTOTALSIZE"), 10, 64)
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_MAXTOTALSIZE to int64, err: %s", err))
			}
			loginfo.isAutoDelete, err = strconv.ParseBool(os.Getenv("CHAINCODE_LOG_ISAUTODELETE"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_ISAUTODELETE to bool, err: %s", err))
			}
			loginfo.daily, err = strconv.ParseBool(os.Getenv("CHAINCODE_LOG_DAILY"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_DAILY to bool, err: %s", err))
			}
			loginfo.rotate, err = strconv.ParseBool(os.Getenv("CHAINCODE_LOG_ROTATE"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_DAILY to bool, err: %s", err))
			}
			loginfo.maxdays, err = strconv.Atoi(os.Getenv("CHAINCODE_LOG_MAXDAYS"))
			if err != nil {
				panic(fmt.Sprintf("Error convert CHAINCODE_LOG_DAILY to int, err: %s", err))
			}
		} else {
			panic(fmt.Sprintln("containerType should not be orderer or peer or chaincode"))
		}

		if loginfo.filepath == "" || len(loginfo.filepath) > FilenameLenMax {

			fmt.Printf("log config args err, filepath:%s, filepath_len:%d, use default arg: %s.\n", loginfo.filepath, len(loginfo.filepath), DefaultFilePath)
			loginfo.filepath = DefaultFilePath
		}

		if loginfo.filename == "" || len(loginfo.filename) > FilepathLenMax {
			fmt.Printf("log config args err,filename:%s,filename_len:%d,use default arg: %s.\n", loginfo.filename, len(loginfo.filename), DefaultFileName)
			loginfo.filename = DefaultFileName
		}

		if loginfo.maxlinesPerFile <= 0 || loginfo.maxlinesPerFile > math.MaxInt32 {
			fmt.Printf("log config args err,maxlinesPerFile:%d,use default arg: %d.\n", loginfo.maxlinesPerFile, DefaultMaxlinesPerFile)
			loginfo.maxlinesPerFile = DefaultMaxlinesPerFile
		}

		if loginfo.maxsizePerFile <= 0 || loginfo.maxsizePerFile > math.MaxInt32 {
			fmt.Printf("log config args err,maxsizePerFile:%d,use default arg: %d.\n", loginfo.maxsizePerFile, DefaultMaxsizePerFile)
			loginfo.maxsizePerFile = DefaultMaxsizePerFile
		}

		if loginfo.maxTotalSize <= 0 || loginfo.maxTotalSize > math.MaxInt64 {
			fmt.Printf("log config args err,maxTotalSize:%d,use default arg: %d.\n", loginfo.maxTotalSize, DefaultMaxTotalSize)
			loginfo.maxTotalSize = DefaultMaxTotalSize
		}

		if loginfo.maxdays <= 0 || loginfo.maxdays > math.MaxInt32 {
			fmt.Printf("log config args err,maxdays:%d,use default arg: %d.\n", loginfo.maxdays, DefaultMaxDays)
			loginfo.maxdays = DefaultMaxDays
		}

		if loglevel < 0 || loglevel > 7 {
			fmt.Printf("log config args err,loglevel:%d,use default arg: %d.\n", loglevel, DefaultLogLevel)
			loglevel = DefaultLogLevel
		}
		fl = GetLogger(loginfo)
	})
	return fl
}

func GetFabricLogger() *FabricLogger {
	return fl
}

func GetLogger(loginfo ggLogInfo) *FabricLogger {
	var separateFile []string

	if loginfo.isOpenGglog == false {
		fmt.Println("use default log.")
		return nil
	}

	os.MkdirAll(loginfo.filepath, 0755)

	l := gglogs.NewLogger(10000)
	l.EnableFuncCallDepth(true)

	for i := LevelEmergency; i <= loglevel; i++ {
		separateFile = append(separateFile, levelNames[i])
	}
	separateFileJson, _ := json.Marshal(separateFile)
	separate := fmt.Sprintf(`"separate":%s`, separateFileJson)

	config := fmt.Sprintf(`"filename":"%s/%s", "maxlines":%d, "maxsize":%d, "maxtotalsize":%d, "daily": %t, "rotate": %t, "maxdays": %d, "isautodelete":%t, `, loginfo.filepath, loginfo.filename, loginfo.maxlinesPerFile, loginfo.maxsizePerFile, loginfo.maxTotalSize, loginfo.daily, loginfo.rotate, loginfo.maxdays, loginfo.isAutoDelete)
	config = "{" + config + separate + "}"

	l.SetLogger(gglogs.AdapterMultiFile, config)
	// default to be 2, because we wrap log with a new method, so adjust the args to 4.
	l.SetLogFuncCallDepth(4)
	fabricLogger := &FabricLogger{
		logger: l,
	}
	return fabricLogger
}

func concat(args ...interface{}) string {
	resultString := fmt.Sprintln(args...)
	// Sprintln will add space between args, and always add an extra '\n' character at the end
	resultString = resultString[0 : len(resultString)-1]
	return resultString
}

func (l *FabricLogger) Debug(v ...interface{}) {
	if loglevel < LevelDebug {
		return
	}
	l.logger.Debug(defaultFormat, concat(v...))
}

func (l *FabricLogger) Debugf(formatString string, v ...interface{}) {
	if loglevel < LevelDebug {
		return
	}
	l.logger.Debug(formatString, v...)
}

func (l *FabricLogger) Info(v ...interface{}) {
	if loglevel < LevelInformational {
		return
	}
	l.logger.Info(defaultFormat, concat(v...))
}

func (l *FabricLogger) Infof(formatString string, v ...interface{}) {
	if loglevel < LevelInformational {
		return
	}
	l.logger.Info(formatString, v...)
}

func (l *FabricLogger) Notice(v ...interface{}) {
	if loglevel < LevelNotice {
		return
	}
	l.logger.Notice(defaultFormat, concat(v...))
}

func (l *FabricLogger) Noticef(formatString string, v ...interface{}) {
	if loglevel < LevelNotice {
		return
	}
	l.logger.Notice(formatString, v...)
}

func (l *FabricLogger) Warning(v ...interface{}) {
	if loglevel < LevelWarning {
		return
	}
	l.logger.Warning(defaultFormat, concat(v...))
}

func (l *FabricLogger) Warningf(formatString string, v ...interface{}) {
	if loglevel < LevelWarning {
		return
	}
	l.logger.Warning(formatString, v...)
}

func (l *FabricLogger) Error(v ...interface{}) {
	if loglevel < LevelError {
		return
	}
	l.logger.Error(defaultFormat, concat(v...))
}

func (l *FabricLogger) Errorf(formatString string, v ...interface{}) {
	if loglevel < LevelError {
		return
	}
	l.logger.Error(formatString, v...)
}

func (l *FabricLogger) Critical(v ...interface{}) {
	if loglevel < LevelCritical {
		return
	}
	l.logger.Critical(defaultFormat, concat(v...))
}

func (l *FabricLogger) Criticalf(formatString string, v ...interface{}) {
	if loglevel < LevelCritical {
		return
	}
	l.logger.Critical(formatString, v...)
}

func (l *FabricLogger) Alert(v ...interface{}) {
	if loglevel < LevelAlert {
		return
	}
	l.logger.Alert(defaultFormat, concat(v...))
}

func (l *FabricLogger) Alertf(formatString string, v ...interface{}) {
	if loglevel < LevelAlert {
		return
	}
	l.logger.Alert(formatString, v...)
}

func (l *FabricLogger) Emergency(v ...interface{}) {
	if loglevel < LevelEmergency {
		return
	}
	l.logger.Emergency(defaultFormat, concat(v...))
}

func (l *FabricLogger) Emergencyf(formatString string, v ...interface{}) {
	if loglevel < LevelEmergency {
		return
	}
	l.logger.Emergency(formatString, v...)
}
