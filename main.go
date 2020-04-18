package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dgl/alertmanager-webhook-signald/signald"
)

var (
	flagListen = flag.String("listen", ":9245", "[ip]:port to listen on for HTTP")
	flagConfig = flag.String("config", "", "YAML configuration filename")

	signalClient *signald.Client
	cfg          *Config
	receivers    = map[string]*Receiver{}
	templates    *template.Template
)

var (
	receivedMetric = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "signald_webhook",
			Subsystem: "alerts",
			Name:      "received_total",
		})
	errorsMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "signald_webhook",
			Subsystem: "alerts",
			Name:      "errors_total",
		}, []string{"type"})

	signaldInfoMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "signald_webhook",
			Subsystem: "signal",
			Name:      "info",
		}, []string{"name", "version"})
	signaldLastKeepaliveMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "signald_webhook",
			Subsystem: "signal",
			Name:      "keepalive",
		})
)

func init() {
	prometheus.Register(receivedMetric)
	prometheus.Register(errorsMetric)
	prometheus.Register(signaldInfoMetric)
	prometheus.Register(signaldLastKeepaliveMetric)
	for _, errorType := range []string{"decode", "handler"} {
		errorsMetric.With(prometheus.Labels{"type": errorType}).Add(0)
	}
}

func main() {
	flag.Parse()
	if *flagConfig == "" {
		flag.Usage()
	}

	var err error
	cfg, err = LoadFile(*flagConfig)
	if err != nil {
		log.Fatal(*flagConfig, ": ", err)
	}

	if len(cfg.Receivers) == 0 {
		log.Fatal(*flagConfig, ": no receivers defined")
	}

	for _, recv := range cfg.Receivers {
		if len(recv.Name) == 0 {
			log.Fatal("Receiver missing 'name:'")
		}
		if _, ok := receivers[recv.Name]; ok {
			log.Fatalf("Duplicate receiver name: %q", recv.Name)
		}
		receivers[recv.Name] = recv
	}

	templates, err = template.FromGlobs(cfg.Templates...)
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	signalClient, err = signald.New()
	if err != nil {
		log.Printf("Error connecting to signald: %v, will attempt to connect later", err)
	}

	prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "signald_webhook",
			Subsystem: "signal",
			Name:      "connected",
			Help:      "True if connected to signald.",
		},
		func() float64 {
			if signalClient.Connected() {
				return 1
			}
			return 0
		},
	))

	// Subscribe if subscribe is true, this helps keep the signal connection alive, even if we don't
	// do anything with the incoming messages.
	subscribe := map[string]bool{}
	for _, recv := range cfg.Receivers {
		if recv.Subscribe != nil && *recv.Subscribe {
			subscribe[recv.Sender] = true
		}
	}
	for user := range subscribe {
		signalClient.Encode(&signald.Subscribe{
			Username: user,
		})
	}

	go handleOutput()
	go keepalive()

	http.HandleFunc("/alert", hook)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*flagListen, nil))
}

// Handles output from signald and deals with reconnection logic
func handleOutput() {
	backoff := 0.0
	for {
		for signalClient.Connected() {
			res, err := signalClient.Decode()
			if err != nil {
				log.Print(err)
				continue
			}

			log.Printf("%T: %v", res, res)
			switch r := res.(type) {
			case *signald.Version:
				signaldInfoMetric.With(
					prometheus.Labels{"name": r.Data["name"], "version": r.Data["version"]}).Set(1)
			case *signald.User:
				signaldLastKeepaliveMetric.Set(float64(time.Now().Unix()))
			}
		}

		time.Sleep(time.Duration(math.Pow(2, backoff)) * time.Second)
		if backoff < 6 {
			backoff += 1
		}
		if err := signalClient.Connect(); err != nil {
			log.Printf("Failed to reconnect: %v", err)
		} else {
			log.Print("Connected to signald")
			backoff = 0
		}
	}
}

func keepalive() {
	if !cfg.Options.KeepAlive {
		return
	}
	for {
		signalClient.Encode(&signald.GetUser{
			Username:        cfg.Defaults.Sender,
			RecipientNumber: cfg.Defaults.Sender,
		})
		time.Sleep(5 * time.Minute)
	}
}

func hook(w http.ResponseWriter, req *http.Request) {
	receivedMetric.Add(1)
	var m Message
	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		log.Printf("Decoding /alert failed: %v", err)
		http.Error(w, "Decode failed", http.StatusBadRequest)
		errorsMetric.With(prometheus.Labels{"type": "decode"}).Add(1)
	}
	err = handle(&m)
	if err != nil {
		log.Print(err)
		http.Error(w, "Handling alert failed", http.StatusInternalServerError)
		errorsMetric.With(prometheus.Labels{"type": "handle"}).Add(1)
	} else {
		fmt.Fprintln(w, "ok")
	}
}

func handle(m *Message) error {
	recv, ok := receivers[m.Receiver]
	if !ok {
		return fmt.Errorf("%q: Receiver not configured", m.Receiver)
	}
	log.Printf("Send via %q: %#v", m.Receiver, recv)

	body, err := templates.ExecuteTextString(recv.Template, m)
	if err != nil {
		body = fmt.Sprintf("%#v: Template expansion failed: %v", m.GroupLabels, err)
	}

	for _, to := range recv.To {
		send := &signald.Send{
			Username:    recv.Sender,
			MessageBody: body,
		}
		if strings.HasPrefix(to, "tel:") {
			send.RecipientNumber = to[4:]
		} else if strings.HasPrefix(to, "group:") {
			send.RecipientGroupID = to[6:]
		} else {
			log.Printf("Unknown to: %q, expected tel:+number or group:id", to)
		}
		if te := signalClient.Encode(send); te != nil {
			err = te
		}
	}
	return err
}
