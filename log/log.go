package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

type CustomArray []interface{}

func (a CustomArray) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, v := range a {
		switch val := v.(type) {
		case string:
			enc.AppendString(val)
		case int:
			enc.AppendInt(val)
		case bool:
			enc.AppendBool(val)
		case float64:
			enc.AppendFloat64(val)
		default:
			err := enc.AppendReflected(val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Event(name string, fields ...interface{}) {
	logger.Info("", zap.String("evt", name), zap.Array("key", CustomArray(fields)))
}
