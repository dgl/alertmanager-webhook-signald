package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
  "net/http"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/template"

	"github.com/dgl/alertmanager-webhook-signald/signald"
)

var (
  flagListen = flag.String("listen", ":9245", "[ip]:port to listen on for HTTP")
	flagConfig = flag.String("config", "", "YAML configuration filename")

	signalClient *signald.Client
	cfg *Config
	receivers = map[string]*Receiver{}
	templates *template.Template
)

func main() {
  flag.Parse()

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

	go logOutput()

  http.HandleFunc("/alert", hook)
  log.Fatal(http.ListenAndServe(*flagListen, nil))
}

func logOutput() {
	for {
		res, err := signalClient.Decode()
		if err != nil {
			log.Print(err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Print(res)
	}
}

func hook(w http.ResponseWriter, req *http.Request) {
  var m Message
  err := json.NewDecoder(req.Body).Decode(&m)
  if err != nil {
    log.Printf("Decoding /alert failed: %v", err)
    http.Error(w, "Decode failed", http.StatusBadRequest)
  }
	err = handle(&m)
	if err != nil {
		log.Print(err)
		http.Error(w, "Handling alert failed", http.StatusInternalServerError)
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
			Username: recv.Sender,
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
