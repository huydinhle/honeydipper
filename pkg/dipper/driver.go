// Copyright 2019 Honey Science Corporation
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, you can obtain one at http://mozilla.org/MPL/2.0/.

package dipper

import (
	"io"
	"os"
	"time"

	"github.com/op/go-logging"
)

// Driver : the helper stuct for creating a honey-dipper driver in golang
type Driver struct {
	RPCCaller
	RPCProvider
	CommandProvider
	Name            string
	Service         string
	State           string
	In              io.Reader
	Out             io.Writer
	Options         interface{}
	MessageHandlers map[string]MessageHandler
	Start           MessageHandler
	Stop            MessageHandler
	Reload          MessageHandler
	ReadySignal     chan bool
}

// NewDriver : create a blank driver object
func NewDriver(service string, name string) *Driver {
	driver := Driver{
		Name:    name,
		Service: service,
		State:   "loaded",
		In:      os.Stdin,
		Out:     os.Stdout,
	}

	driver.RPCProvider.Init("rpc", "return", driver.Out)
	driver.RPCCaller.Init(&driver, "rpc", "call")
	driver.CommandProvider.Init("eventbus", "return", driver.Out)

	driver.MessageHandlers = map[string]MessageHandler{
		"command:options":  driver.ReceiveOptions,
		"command:ping":     driver.Ping,
		"command:start":    driver.start,
		"command:stop":     driver.stop,
		"rpc:call":         driver.RPCProvider.Router,
		"rpc:return":       driver.HandleReturn,
		"eventbus:command": driver.CommandProvider.Router,
	}

	driver.GetLogger()
	return &driver
}

// Run : start a loop to communicate with daemon
func (d *Driver) Run() {
	Logger.Infof("[%s] driver loaded", d.Service)
	for {
		func() {
			defer SafeExitOnError("[%s] Resuming driver message loop", d.Service)
			defer CatchError(io.EOF, func() {
				Logger.Fatalf("[%s] daemon closed channel", d.Service)
			})
			for {
				msg := FetchRawMessage(d.In)
				go func() {
					defer SafeExitOnError("[%s] Continuing driver message loop", d.Service)
					if handler, ok := d.MessageHandlers[msg.Channel+":"+msg.Subject]; ok {
						handler(msg)
					} else {
						Logger.Infof("[%s] skipping message without handler: %s:%s", d.Service, msg.Channel, msg.Subject)
					}
				}()
			}
		}()
	}
}

// Ping : respond to daemon ping request with driver state
func (d *Driver) Ping(msg *Message) {
	d.SendMessage(&Message{
		Channel: "state",
		Subject: d.State,
	})
}

// ReceiveOptions : receive options from daemon
func (d *Driver) ReceiveOptions(msg *Message) {
	msg = DeserializePayload(msg)
	Recursive(msg.Payload, RegexParser)
	d.Options = msg.Payload
	Logger = nil
	d.GetLogger()
	d.ReadySignal <- true
}

func (d *Driver) start(msg *Message) {
	select {
	case <-d.ReadySignal:
	case <-time.After(time.Second):
	}

	if d.State == "alive" {
		if d.Reload != nil {
			d.Reload(msg)
		} else {
			d.State = "cold"
		}
	} else {
		if d.Start != nil {
			d.Start(msg)
		}
		d.State = "alive"
	}
	d.Ping(msg)
}

func (d *Driver) stop(msg *Message) {
	d.State = "exit"
	if d.Stop != nil {
		d.Stop(msg)
	}
	d.Ping(msg)
	Logger.Fatalf("[%s] quiting on daemon request", d.Service)
}

// SendMessage : send a prepared message to daemon
func (d *Driver) SendMessage(m *Message) {
	Logger.Infof("[%s] sending raw message to daemon %s:%s", d.Service, m.Channel, m.Subject)
	SendMessage(d.Out, m)
}

// GetOption : get the data from options map with the key
func (d *Driver) GetOption(path string) (interface{}, bool) {
	return GetMapData(d.Options, path)
}

// GetOptionStr : get the string data from options map with the key
func (d *Driver) GetOptionStr(path string) (string, bool) {
	return GetMapDataStr(d.Options, path)
}

// we have to keep hold of the os.File object to
// avoid being closed by garbage collector (runtime.setFinalizer)
var logFile *os.File

// GetLogger : getting a logger for the driver
func (d *Driver) GetLogger() *logging.Logger {
	if Logger == nil {
		levelstr, ok := d.GetOptionStr("data.loglevel")
		if !ok {
			levelstr = "INFO"
		}
		if logFile == nil {
			logFile = os.NewFile(uintptr(3), "log")
		}
		return GetLogger(d.Name, levelstr, logFile)
	}
	return Logger
}

// GetStream getting a output stream for a feature
func (d *Driver) GetStream(feature string) io.Writer {
	return d.Out
}

// GetName returns the name of the driver
func (d *Driver) GetName() string {
	return d.Name
}
