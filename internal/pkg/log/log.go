package log

import (
	"context"
	"fmt"
	"log"
	"ryzenlo/to2cloud/configs"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

var levelDict = map[string]logrus.Level{
	"trace": logrus.TraceLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
}

func InitLogger(ctx context.Context, conf *configs.Config) {
	//
	if _, err := os.Stat(conf.Log.DirPath); os.IsNotExist(err) {
		//create
		err = os.MkdirAll(conf.Log.DirPath, os.ModePerm)
		if err != nil {
			log.Fatalf("cannot create diretory for logger,%v", err)
		}
	}
	//
	logFileName := conf.Log.DirPath + "/" + conf.Log.FileName
	//
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open log file %s ,%v", logFileName, err)
	}
	//
	aLogger := logrus.New()
	var usedLevel logrus.Level
	if _, ok := levelDict[conf.Log.Level]; !ok {
		usedLevel = logrus.InfoLevel
	} else {
		usedLevel = levelDict[conf.Log.Level]
	}
	aLogger.SetLevel(usedLevel)
	aLogger.SetFormatter(&logrus.JSONFormatter{})
	aLogger.SetOutput(f)
	Logger = aLogger
	go monitorLogFile(ctx, conf, f)
}

func monitorLogFile(ctx context.Context, conf *configs.Config, lastOpenFile *os.File) {
	Logger.Infoln("Start monitoring log file.")
	monitorTimer := time.NewTimer(time.Minute)
	for {
		select {
		case <-ctx.Done():
			Logger.Infoln("Stop monitoring log file.")
			lastOpenFile.Close()
			Logger.Out = os.Stdout
			return
		case <-monitorTimer.C:
			Logger.Infoln("Monitoring log file...")
			succeed := monitorTimer.Reset(time.Minute)
			if !succeed {
				monitorTimer = time.NewTimer(time.Minute)
			}
			if !checkLogFileExceedMaxSize(conf, lastOpenFile) {
				continue
			}
			newFileName := fmt.Sprintf("%s/%s_%s", conf.Log.DirPath, conf.Log.FileName, time.Now().Format("20060102150405"))
			//
			newOpenFile, err := os.OpenFile(newFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				Logger.Errorln("cannot open log file %s ,%v", newFileName, err)
				continue
			}
			Logger.SetOutput(newOpenFile)
			lastOpenFile.Close()
			lastOpenFile = newOpenFile
		}
	}
}
func checkLogFileExceedMaxSize(conf *configs.Config, f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	if conf.Log.MaxSize == 0 {
		return false
	}
	if stat.Size() > conf.Log.MaxSize {
		return true
	}
	return false
}
