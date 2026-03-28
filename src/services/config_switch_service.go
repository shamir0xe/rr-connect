package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type configSwitchStruct struct {
	placeholder string
	router      RouterInterface
	cfg         *viper.Viper
}

type ConfigSwitchInterface interface {
	Run(context.Context, *sync.WaitGroup, <-chan bool) error
}

func NewConfigSwitchService(cfg *viper.Viper, router RouterInterface) (ConfigSwitchInterface, error) {
	sub := cfg.Sub("config-switch")
	return &configSwitchStruct{
		router:      router,
		placeholder: sub.GetString("placeholder"),
		cfg:         cfg,
	}, nil
}

func (cs *configSwitchStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan <-chan bool) error {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("ConfigSwitch -> END")
			return nil
		case <-triggerChan:

			err := cs.makeConfig()
			if err != nil {
				return err
			}

			err = cs.postProcess()
			if err != nil {
				return err
			}

			go cs.notify(ctx)
		}
	}
}

func (cs *configSwitchStruct) makeConfig() error {
	templateCfg := cs.cfg.GetString("config-switch.template-cfg")
	outputCfg := cs.cfg.GetString("config-switch.output-cfg")

	log.Printf("ConfigSwitch -> creating new config [%s]\n", cs.router.Pick())
	input, err := os.Open(templateCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	output, err := os.Create(outputCfg)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(
		"sed",
		fmt.Sprintf("s|%s|%s|g", cs.placeholder, cs.router.Next()),
	)
	cmd.Stdin = input
	cmd.Stdout = output
	err = cmd.Run()
	return err
}

func (cs *configSwitchStruct) postProcess() error {
	sub := cs.cfg.Sub("config-switch.post-process")

	if !sub.GetBool("enabled") {
		return nil
	}

	command := sub.GetString("command")
	args := sub.GetStringSlice("args")
	help := sub.GetString("help")

	log.Printf("ConfigSwitch -> post-process %s\n", help)

	err := exec.Command(command, args...).Run()
	return err
}

func (cs *configSwitchStruct) notify(ctx context.Context) error {
	sub := cs.cfg.Sub("config-switch.notify")

	if !sub.GetBool("enabled") {
		return nil
	}

	command := sub.GetString("command")
	notifyArgs := sub.GetStringSlice("args")
	delay := sub.GetDuration("delay")
	timeout := sub.GetDuration("timeout")

	log.Printf("ConfigSwitch -> executing notify command after %s delay\n", delay)
	timer := time.NewTimer(delay)
	<-timer.C

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var args []string = make([]string, len(notifyArgs))
	for i, arg := range notifyArgs {
		if strings.Contains(arg, cs.placeholder) {
			args[i] = strings.ReplaceAll(arg, cs.placeholder, cs.router.Previous())
		} else {
			args[i] = arg
		}
	}

	err := exec.CommandContext(timeoutCtx, command, args...).Run()
	if err != nil {
		log.Printf("ConfigSwitch -> notify command failed: %v\n", err)
	} else {
		log.Printf("ConfigSwitch -> notified successfully [%s]!\n", cs.router.Previous())
	}

	return err
}
