package logger

// String 类似 zap.String()
func String(key string, val any) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}