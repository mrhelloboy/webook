package logger

// String 类似 zap.String()
func String(key string, val any) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Int32(key string, val int32) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Int64(key string, val int64) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
