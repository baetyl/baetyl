package main

import (
	"encoding/binary"
	"io"
	"os"
	"path"
	"path/filepath"

	sdk "github.com/baidu/openedge/sdk/go"
	"github.com/baidu/openedge/utils"
)

func main() {
	sdk.Run(func(ctx openedge.Context) error {
		var cfg Config
		err := utils.LoadYAML("etc/openedge/service.yml", &cfg)
		if err != nil {
			return err
		}
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return err
		}
		cdir := path.Dir(path.Dir(exe))
		ir, iw, err := os.Pipe()
		if err != nil {
			return err
		}
		defer iw.Close()
		or, ow, err := os.Pipe()
		if err != nil {
			ir.Close()
			return err
		}
		defer or.Close()
		p, err := os.StartProcess(
			path.Join(cdir, "bin", "runtime"),
			[]string{
				"openedge-function-python27-runtime",
				cfg.Name,
				cfg.Handler,
			},
			&os.ProcAttr{
				Files: []*os.File{ir, ow, os.Stderr},
			},
		)
		ir.Close()
		ow.Close()
		if err != nil {
			return err
		}
		err = ctx.Subscribe(
			openedge.TopicInfo{
				Topic: cfg.Subscribe.Topic,
				QoS:   cfg.Subscribe.QoS,
			}, func(msg *openedge.Message) error {
				err := send(iw, msg)
				if err != nil {
					openedge.Infoln("failed to send to runtime:", err.Error())
					return nil
				}
				reply, err := recv(or)
				if err != nil {
					openedge.Infoln("failed to recv from runtime:", err.Error())
					return nil
				}
				if len(cfg.Publish.Topic) != 0 {
					err = ctx.SendMessage(&openedge.Message{
						Topic:   cfg.Publish.Topic,
						QoS:     0, // FIXME cfg.Publish.QoS,
						Payload: reply,
					})
					if err != nil {
						openedge.Warnln("failed to send message:", err.Error())
					}
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
		ctx.Wait()
		p.Kill()
		p.Wait()
		return nil
	})
}

func send(w io.Writer, msg *openedge.Message) error {
	err := binary.Write(w, binary.BigEndian, uint16(len(msg.Topic)))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg.Topic))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg.QoS)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, uint32(len(msg.Payload)))
	if err != nil {
		return err
	}
	_, err = w.Write(msg.Payload)
	if err != nil {
		return err
	}
	return nil
}

func recv(r io.Reader) ([]byte, error) {
	var len uint32
	err := binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, int(len))
	_, err = r.Read(buf)
	return buf, err
}
