package services

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Placeholders struct {
	extra  string
	router string
}

type configSwitchStruct struct {
	router       RouterInterface
	cfg          *viper.Viper
	runCmd       func(ctx context.Context, name string, args ...string) *exec.Cmd
	placeholders *Placeholders
}

type ConfigSwitchInterface interface {
	Run(context.Context, *sync.WaitGroup, <-chan bool) error
}

func NewConfigSwitchService(cfg *viper.Viper, router RouterInterface) (ConfigSwitchInterface, error) {
	sub := cfg.Sub("config-switch")
	return &configSwitchStruct{
		router: router,
		placeholders: &Placeholders{
			router: sub.GetString("placeholders.router"),
			extra:  sub.GetString("placeholders.extra"),
		},
		cfg:    sub,
		runCmd: exec.CommandContext,
	}, nil
}

func (cs *configSwitchStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan <-chan bool) error {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("config-switch -> END")
			return nil
		case <-triggerChan:

			routerName, err := cs.makeConfig(ctx)
			if err != nil {
				log.Printf("config-switch -> make-config -> error: %v", err)
				continue
			}

			err = cs.postProcess(ctx)
			if err != nil {
				log.Printf("config-switch -> post-process -> error: %v", err)
			}

			go cs.notify(ctx, routerName)
		}
	}
}

func (cs *configSwitchStruct) makeConfig(_ context.Context) (string, error) {
	templateCfg := cs.cfg.GetString("template-cfg")
	outputCfg := cs.cfg.GetString("output-cfg")

	selectedRouter := cs.router.Next()
	log.Printf("config-switch -> creating new config [%s]\n", selectedRouter)

	content, err := os.ReadFile(templateCfg)
	if err != nil {
		return "", err
	}

	replaced := strings.ReplaceAll(string(content), cs.placeholders.router, selectedRouter)

	if err := os.WriteFile(outputCfg, []byte(replaced), 0644); err != nil {
		return "", err
	}

	return selectedRouter, nil
}

func (cs *configSwitchStruct) postProcess(ctx context.Context) error {
	sub := cs.cfg.Sub("post-process")

	if !sub.GetBool("enabled") {
		return nil
	}

	command := sub.GetString("command")
	args := sub.GetStringSlice("args")
	help := sub.GetString("help")
	timeout := sub.GetDuration("timeout")

	log.Printf("config-switch -> post-process %s\n", help)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := cs.runCmd(timeoutCtx, command, args...)
	err := cmd.Run()

	log.Println("config-switch -> post-process DONE")
	return err
}

func (cs *configSwitchStruct) notify(ctx context.Context, routerName string) error {
	sub := cs.cfg.Sub("notify")

	if !sub.GetBool("enabled") {
		return nil
	}

	command := sub.GetString("command")
	notifyArgs := sub.GetStringSlice("args")
	delay := sub.GetDuration("delay")
	timeout := sub.GetDuration("timeout")
	extraPath := sub.GetString("extra-path")

	extraResult := ""
	if extraPath != "" {
		extraFile, err := os.ReadFile(extraPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("config-switch -> failed to read extra-path %q: %v\n", extraPath, err)
			}
		} else {
			extraResult = strings.TrimSpace(string(extraFile))
		}
	}

	log.Printf("config-switch -> executing notify command after %s delay\n", delay)
	timer := time.NewTimer(delay)
	<-timer.C

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var args []string = make([]string, len(notifyArgs))
	for i, arg := range notifyArgs {
		args[i] = arg
		if strings.Contains(args[i], cs.placeholders.router) {
			args[i] = strings.ReplaceAll(args[i], cs.placeholders.router, routerName)
		}
		if extraResult != "" && strings.Contains(args[i], cs.placeholders.extra) {
			args[i] = strings.ReplaceAll(args[i], cs.placeholders.extra, extraResult)
		}
	}

	err := cs.runCmd(timeoutCtx, command, args...).Run()
	if err != nil {
		log.Printf("config-switch -> notify command failed: %v\n", err)
	} else {
		log.Printf("config-switch -> notified successfully [%s]!\n", routerName)
	}

	return err
}
