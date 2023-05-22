package gotelem

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kschamplin/gotelem/skylab"
)

// CanWriter
type CanWriter struct {
	output *os.File
	cd candumpJSON
	jsonBuf []byte
}

// send writes the frame to the file.
func (cw *CanWriter) Send(f *Frame) (err error) {
	cw.cd.Timestamp = float64(time.Now().Unix())

	cw.cd.Id = uint64(f.Id)

	cw.cd.Data, err = skylab.FromCanFrame(f.Id, f.Data)

	if err != nil {
		return
	}
	out, err := json.Marshal(cw.cd)
	if err != nil {
		return
	}
	fmt.Fprintln(cw.output, string(out))
	return err
}

func (cw *CanWriter) Close() error {
	return cw.output.Close()
}

func OpenCanWriter(name string) (*CanWriter, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	cw := &CanWriter{
		output: f,
	}
	return cw, nil
}

type candumpJSON struct {
	Timestamp float64       `json:"ts"`
	Id        uint64        `json:"id"`
	Data      skylab.Packet `json:"data"`
}
