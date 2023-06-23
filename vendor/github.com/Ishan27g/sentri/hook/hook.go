package hook

import (
	"context"
	"io"

	"github.com/Ishan27g/sentri/internal"
)

func Hook(appName string, w io.Writer) io.Writer {
	internal.InitPool(context.Background())
	pr, pw := io.Pipe()
	internal.Ready(appName, pr)
	w = io.MultiWriter(w, pw)
	return w
}
