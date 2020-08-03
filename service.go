// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package videodir

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

var elog debug.Log

type MyService struct{
	Name string
	Desc string
	Debug bool
}

func NewService(name, desc string, isInteractive bool) *MyService {
	return &MyService{
		Name: name,
		Desc: desc,
		Debug: isInteractive,
	}
}

func (service *MyService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// init and run app server in the noninteractive mode!!!
	// Working directory for windows - exe directory!!
	ex, err := os.Executable()
	if err != nil {
		elog.Error(1, fmt.Sprintf("get executable %v", err))
		return
	}
	workDir := filepath.Dir(ex)
	logger, err := NewLogger(workDir, service.Debug)
	if err != nil {
		elog.Error(1, fmt.Sprintf("error init logger %v", err))
		return
	}
	app := NewApp(filepath.Join(filepath.Dir(ex), CONF_FILE), logger)
	go app.Serve()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				//testOutput := strings.Join(args, "-")
				//testOutput += fmt.Sprintf("-%d", c.Context)
				//elog.Info(1, testOutput)
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func (service *MyService) Run() error {
	var err error
	if service.Debug {
		elog = debug.New(service.Name)
	} else {
		elog, err = eventlog.Open(service.Name)
		if err != nil {
			return err
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service %v", service.Name, service.Debug))
	run := svc.Run
	if service.Debug {
		run = debug.Run
	}
	//err = run(service.Name, &MyService{})
	err = run(service.Name, service)
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", service.Name, err))
		return err
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", service.Name))
	return nil
}

func (service *MyService) Install() error {
	exepath, err := os.Executable()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(service.Name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", service.Name)
	}
	s, err = m.CreateService(service.Name, exepath, mgr.Config{DisplayName: service.Desc}, "is", "auto-started")
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(service.Name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func (service *MyService) Remove() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(service.Name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", service.Name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(service.Name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func (service *MyService) Start() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(service.Name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func (service *MyService) Control(c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(service.Name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

