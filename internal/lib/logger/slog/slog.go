package slog

import "log/slog"

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Int64(s string, id int64) slog.Attr {
	return slog.Attr{
		Key:   s,
		Value: slog.Int64Value(id),
	}
}
