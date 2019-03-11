package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/sfreiberg/gotwilio"
)

type config struct {
	AccountSID       string   `yaml:"account-sid"`
	AuthToken        string   `yaml:"auth-token"`
	Subscribers      []string `yaml:"subscribers"`
	NotificationText string   `yaml:"notification-text"`
	Number           string   `yaml:"number"`
}

func (conf config) Validate() error {
	missing := "missing value: %s"

	if conf.AccountSID == "" {
		return fmt.Errorf(missing, "account SID")
	} else if conf.AuthToken == "" {
		return fmt.Errorf(missing, "auth token")
	} else if len(conf.Subscribers) == 0 {
		return fmt.Errorf(missing, "subscribers")
	} else if conf.NotificationText == "" {
		return fmt.Errorf(missing, "notification text")
	} else if conf.Number == "" {
		return fmt.Errorf(missing, "number")
	}

	return nil
}

func loadConfig() (config, error) {
	conf := config{}

	number := flag.String("from", "", "The number from which the SMS will be sent")
	subs := flag.String("subs", "", "Subscribers list")
	text := flag.String("text", "", "Text body")
	sid := flag.String("sid", "", "Twilio account SID")
	token := flag.String("token", "", "Twilio account token")
	confPath := flag.String("config", "", "Config path")
	flag.Parse()

	if *confPath != "" {
		f, err := ioutil.ReadFile(*confPath)
		if err != nil {
			return conf, err
		}
		if err = yaml.Unmarshal(f, &conf); err != nil {
			return conf, errors.Wrap(err, "unmarshal config")
		}
	}

	if *sid != "" {
		conf.AccountSID = *sid
	}
	if *token != "" {
		conf.AuthToken = *token
	}
	if *subs != "" {
		conf.Subscribers = strings.Split(*subs, ",")
	}
	if *text != "" {
		conf.NotificationText = *text
	}
	if *number != "" {
		conf.Number = *number
	}

	if err := conf.Validate(); err != nil {
		return conf, errors.Wrap(err, "invalid config")
	}

	return conf, nil
}

func propagate(twilio *gotwilio.Twilio, subs []string, from, text string) error {
	for _, sub := range subs {
		log.Printf("sending notification to %s from %s. body: %q", sub, from, text)
		_, exc, err := twilio.SendSMS(from, sub, text, "", "")
		if err != nil {
			return errors.Wrap(err, "send SMS")
		}
		if exc != nil {
			return fmt.Errorf("send SMS: code: %d, status: %d, message: %s, more info: %s", exc.Code, exc.Status, exc.Message, exc.MoreInfo)
		}
	}

	return nil
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %s", err)
	}

	twilio := gotwilio.NewTwilioClient(conf.AccountSID, conf.AuthToken)

	if err := propagate(twilio, conf.Subscribers, conf.Number, conf.NotificationText); err != nil {
		log.Fatal(err)
	}
}
