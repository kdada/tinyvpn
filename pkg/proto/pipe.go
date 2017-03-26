package proto

import (
	"io"
	"log"
)

// Pipe pipes r to w
func Pipe(name string, running *bool, r io.Reader, w io.Writer) <-chan struct{} {
	signal := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		for *running {
			rc, err := r.Read(buf)
			if err != nil {
				log.Println(name, "pipe read error", err)
				break
			}
			wc, err := w.Write(buf[:rc])
			if err != nil {
				log.Println(name, "pipe write error", err)
				break
			}
			if rc != wc {
				log.Println(name, "broken pipe", "read count:", rc, "write count:", wc)
				break
			}
		}
		log.Println(name, "pip stop")
		signal <- struct{}{}
	}()
	return signal
}
