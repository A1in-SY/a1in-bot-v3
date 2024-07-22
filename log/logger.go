package log

import (
	"a1in-bot-v3/log/conf"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logo = `
      __        ____     __    _____  ___-  _______     ______  ___________- ___      ___  _______  
     /""\      /  " \   |" \  (\"   \|"  \ |   _  "\   /    " \("     _   ")|"  \    /"  |/" __   ) 
    /    \    /__|| |   ||  | |.\\   \    |(. |_)  :) // ____  \)__/  \\__/  \   \  //  /(__/ _) ./ 
   /' /\  \      |: |   |:  | |: \.   \\  ||:     \/ /  /    ) :)  \\_ /      \\  \/. ./     /  //  
  //  __'  \    _\  |   |.  | |.  \    \. |(|  _  \\(: (____/ //   |.  |       \.    //   __ \_ \\  
 /   /  \\  \  /" \_|\  /\  |\|    \    \ ||: |_)  :)\        /    \:  |        \\   /   (: \__) :\ 
(___/    \___)(_______)(__\_|_)\___|\____\)(_______/  \"_____/      \__|         \__/     \_______) 
                                                                                                    
`
)

func InitLogger(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init log util, an error happen: %v", err.Error())
		}
	}()
	c := conf.DefaultConf()
	err = toml.Unmarshal(cbs, c)
	if err != nil {
		return
	}
	err = c.Check()
	if err != nil {
		return
	}

	var (
		encoder     zapcore.Encoder
		writeSyncer zapcore.WriteSyncer
	)
	lumberJackLogger := &lumberjack.Logger{
		Filename:   c.LogConf.FileName,
		MaxSize:    c.LogConf.MaxSize,
		MaxAge:     c.LogConf.MaxAge,
		MaxBackups: c.LogConf.MaxBackups,
		Compress:   false,
	}
	if c.LogConf.IsStdout {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
	} else {
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	}
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02-15:04:05.000"))
	}
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoder = zapcore.NewConsoleEncoder(encodeConfig)

	level := new(zapcore.Level)
	err = level.UnmarshalText([]byte(c.LogConf.Level))
	if err != nil {
		return
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	var logger *zap.Logger
	if c.LogConf.IsStackTrace {
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		logger = zap.New(core, zap.AddCaller())
	}
	zap.ReplaceGlobals(logger)
	logger.Info(logo)
	return
}
