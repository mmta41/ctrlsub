package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	HostsList        string
	Threads          int
	Debug            bool
	ForceHTTPS       bool
	FallbackProtocol bool
	Timeout          int
	OutputFile       string
	Quote            bool
	QuoteEmpty       bool
	Colorize         bool
	JSON             bool
	JSONPretty       bool
	Provider         string
	Help             bool
}

var config = Config{}

func main() {
	parseArguments()
	configLogger()
	err := InitializeProviders()
	if err != nil {
		log.Fatal(err)
	}

	targets := strings.Split(config.HostsList, ",")

	hosts := make(chan string, 0)

	wg := sync.WaitGroup{}
	wg.Add(config.Threads)

	for config.Threads > 0 {
		config.Threads -= 1
		go func() {
			for {
				host := <-hosts
				if host == "" {
					break
				}
				state := &LogFormat{
					Status:   false,
					Target:   host,
					Cname:    "",
					Provider: "",
				}
				log.WithFields(state.toField(nil)).Debugln("Testing")
				Checker(host, state)
			}
			wg.Done()
		}()
	}

	for _, h := range targets {
		hosts <- h
	}
	close(hosts)
	wg.Wait()
}

func Checker(target string, state *LogFormat) {
	TargetCNAME, err := net.LookupCNAME(target)
	if err != nil {
		log.WithFields(state.Done().toField(err)).Debugln("no cname")
		return
	}
	state.Cname = TargetCNAME
	if p := FindProvider(TargetCNAME); p != nil {
		state.Provider = p.Name
		log.WithFields(state.toField(nil)).Debugln("Selected")
		Check(target, p, state)
	} else {
		log.WithFields(state.Done().toField(nil)).Println("no provider")
	}
}

func Check(target string, p *Provider, state *LogFormat) {
	_, body, err := Request(target, time.Duration(config.Timeout)*time.Second, config.ForceHTTPS, config.FallbackProtocol)
	if err != nil {
		log.WithFields(state.Done().toField(err)).Debugln("testing target")
		return
	}

	for _, resp := range p.Response {
		if strings.Contains(string(body), resp) {
			state.Status = true
			if p.Name == "cloudfront" && !config.ForceHTTPS {
				_, bytes, err := Request(target, 120, true, false)
				if err != nil {
					state.Status = false
					log.WithFields(state.Done().toField(err)).Debugln("cloudfront no source at https")
					log.WithFields(state.toField(nil)).Println("cloudfront no source at https")
					return
				}
				if !strings.Contains(string(bytes), resp) {
					state.Status = false
					log.WithFields(state.Done().toField(nil)).Println("not matched")
					return
				}
			}
			log.WithFields(state.Done().toField(nil)).Println("matched")
			return
		}
	}
	log.WithFields(state.Done().toField(nil)).Println("matched")
}

func parseArguments() {
	flag.IntVar(&config.Threads, "t", 20, "Number of threads to use")
	flag.BoolVar(&config.Debug, "v", false, "Debug mode")
	flag.StringVar(&config.HostsList, "u", "", "Comma separated hosts to check takeovers on")
	flag.BoolVar(&config.FallbackProtocol, "fb", true, "Fallback protocol http <=> https vice versa")
	flag.BoolVar(&config.ForceHTTPS, "https", false, "Force HTTPS connections")
	flag.BoolVar(&config.JSON, "j", false, "Output format as json")
	flag.BoolVar(&config.JSONPretty, "jp", false, "Output format as json with pretty print")
	flag.BoolVar(&config.Quote, "qa", false, "Quote all fields")
	flag.BoolVar(&config.QuoteEmpty, "qe", false, "Quote empty fields")
	flag.BoolVar(&config.Colorize, "c", false, "Colorize output")
	flag.IntVar(&config.Timeout, "timeout", 10, "Seconds to wait before timeout.")
	flag.StringVar(&config.OutputFile, "o", "", "File to write enumeration output to")
	flag.StringVar(&config.Provider, "p", "providers.json", "Load providers from a file")
	flag.BoolVar(&config.Help, "h", false, "show this page")
	flag.Parse()

	if config.Help {
		flag.Usage()
		os.Exit(2)
	}
	validateArguments()
}

func validateArguments() {
	required := []string{"u"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			log.Fatalf("missing required -%s argument/flag\n", req)
			flag.Usage()
			os.Exit(2)
		}
	}
}

func configLogger() {
	log.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
		},
	})
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    !config.Colorize,
		ForceQuote:       config.Quote,
		DisableTimestamp: true,
		SortingFunc:      SortLog,
		QuoteEmptyFields: config.QuoteEmpty,
	})
	if config.JSON || config.JSONPretty {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:  "",
			PrettyPrint:      config.JSONPretty,
			DisableTimestamp: true,
		})
	}

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}
}
