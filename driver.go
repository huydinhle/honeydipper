package main

import (
	"github.com/honeyscience/honeydipper/dipper"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// NewDriver : create a driver object to run the program process
func NewDriver(data map[string]interface{}) Driver {
	cmd, ok := data["Executable"].(string)
	if !ok {
		cmd = ""
	}

	args, ok := data["Arguments"].([]string)
	if !ok {
		args = []string{}
	}

	driver := Driver{
		Executable: cmd,
		Arguments:  args,
	}
	return driver
}

func (runtime *DriverRuntime) start(service string) {
	runtime.service = service

	if runtime.driver.PreStart != nil {
		runtime.driver.PreStart(service, runtime)
	}

	args := append([]string{service}, runtime.driver.Arguments...)
	run := exec.Command(runtime.driver.Executable, args...)
	if input, err := run.StdoutPipe(); err != nil {
		log.Panicf("[%s] Unable to link to driver stdout %v", service, err)
	} else {
		runtime.input = input
		runtime.readfd = int(input.(*os.File).Fd())
		log.Printf("[%s] driver %s input fd %v", service, runtime.meta.Name, runtime.readfd)
		flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(runtime.readfd), syscall.F_GETFL, 0)
		if errno != 0 {
			panic(errno.Error())
		}
		flags |= syscall.O_NONBLOCK
		_, _, errno = syscall.Syscall(syscall.SYS_FCNTL, uintptr(runtime.readfd), syscall.F_SETFL, uintptr(flags))
		if errno != 0 {
			panic(errno.Error())
		}
	}
	if output, err := run.StdinPipe(); err != nil {
		log.Panicf("[%s] Unable to link to driver stdin %v", service, err)
	} else {
		runtime.output = output
	}
	run.Stderr = os.Stderr
	if err := run.Start(); err != nil {
		log.Panicf("[%s] Failed to start driver %v", service, err)
	}

	runtime.Run = run
	runtime.sendOptions()
}

func (runtime *DriverRuntime) sendOptions() {
	runtime.sendMessage(&dipper.Message{
		Channel: "command",
		Subject: "options",
		IsRaw:   false,
		Payload: runtime.data,
	})
	runtime.sendMessage(&dipper.Message{
		Channel: "command",
		Subject: "start",
	})
}

func (runtime *DriverRuntime) sendMessage(msg *dipper.Message) {
	if msg.IsRaw && msg.Payload != nil {
		dipper.SendRawMessage(runtime.output, msg.Channel, msg.Subject, msg.Payload.([]byte))
	} else {
		dipper.SendMessage(runtime.output, msg.Channel, msg.Subject, msg.Payload)
	}
}

func (runtime *DriverRuntime) fetchMessages() (messages []*dipper.Message) {
	defer dipper.SafeExitOnError(
		"failed to fetching messages from driver %s.%s",
		runtime.service,
		runtime.meta.Name,
	)
	func() {
		defer dipper.IgnoreError(io.EOF)
		for {
			message := dipper.FetchMessage(runtime.input)
			log.Printf("[%s-%s] driver fetched message %+v", runtime.service, runtime.meta.Name, *message)
			messages = append(messages, message)
		}
	}()

	return messages
}
