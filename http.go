package main

import (
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AlbinoDrought/creamy-gateway-override/remote"
)

const rawTemplateViewGateways = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>Creamy Gateway Override</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style type="text/css">
		html, body {
			font-family: mono;
			background-color: #1b1b1b;
			color: #ababab;
		}
		form {
			display: flex;
			flex-direction: row;
			align-items: center;
		}
		.gateways {
			display: flex;
			flex-direction: column;
			justify-content: center;
		}
		@media(min-width: 400px) {
			.gateways {
				max-width: 400px;
			}
		}

		.gateway {
			padding: 1em;
			margin: 1em;
			border: 1px solid rgba(0,0,0,0.5);
			
			display: flex;
			text-align: center;
			flex-direction: column;
			justify-content: center;
			align-items: center;
		}
		@media(min-width: 400px) {
			.gateway {
				text-align: unset;
				flex-direction: row;
				align-items: flex-start;
				justify-content: space-between;
			}
		}

		.gateway--active {
			background-color: #00550055;
		}

		.gateway__status {
			display: flex;
			flex-direction: column;
			justify-content: center;
			align-items: center;
		}

		.status--online { color: lawngreen; }
		.status--offline { color: crimson; }
		</style>
	</head>
	<body>
		<p>Hello <strong>{{ .Source }}</strong></p>

		<div class="gateways">
			{{ range $element := .Gateways }}
				{{ if (eq $element.Active true) }}
					<div class="gateway gateway--active">
						<strong>{{ $element.Label }}</strong>

						{{ if (eq $element.HasKnownStatus true) }}
						<div class="gateway__status">
							{{ if (eq $element.Online true) }}
							<span class="status status--online">Online</span>
							{{ else }}
							<span class="status status--offline">Offline</span>
							{{ end }}

							<span>{{ $element.RoundtripTime }}</span>
						</div>
						{{ end }}

						<span>(active)</span>
					</div>
				{{ else }}
					<div class="gateway gateway--inactive">
						<span>{{ $element.Label }}</span>

						{{ if (eq $element.HasKnownStatus true) }}
						<div class="gateway__status">
							{{ if (eq $element.Online true) }}
							<span class="status status--online">Online</span>
							{{ else }}
							<span class="status status--offline">Offline</span>
							{{ end }}
							
							<span>{{ $element.RoundtripTime }}</span>
						</div>
						{{ end }}

						<form method="POST">
							<button type="submit" name="gateway" value="{{ $element.Name }}">Activate</button>
						</form>
					</div>
				{{ end }}
			{{ end }}
		</div>
	</body>
</html>
`

var templateViewGateways = template.Must(template.New("viewGateways").Parse(rawTemplateViewGateways))

func getSource(r *http.Request) (string, error) {
	if cfg.TrustForwardedHeaders {
		forwardedAddresses := r.Header.Get("X-Forwarded-For")

		if forwardedAddresses != "" {
			firstComma := strings.Index(forwardedAddresses, ",")
			if firstComma == -1 {
				return forwardedAddresses, nil
			}

			return forwardedAddresses[:firstComma], nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	return ip, err
}

func handlerViewGateways(w http.ResponseWriter, r *http.Request) {
	gateways := cfg.Gateways
	activeGatewayName := deleteDork

	ip, err := getSource(r)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("could not get source"))
		return
	}

	gatewayStatus, err := getGatewayStatus()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	gatewayStatusMap := make(map[string]remote.Gateway, len(gatewayStatus))
	for _, gateway := range gatewayStatus {
		gatewayStatusMap[gateway.Name()] = gateway
	}

	activeRule, err := getActiveRule(cfg.RemoteInterface, ip)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	if activeRule != nil {
		activeGatewayName = activeRule.Gateway()
	}

	gatewaysWithState := make([]struct {
		Name   string
		Label  string
		Active bool

		HasKnownStatus bool
		RoundtripTime  string
		Online         bool
	}, len(gateways))
	for i, gateway := range gateways {
		gatewaysWithState[i].Name = gateway.Name
		gatewaysWithState[i].Label = gateway.Label
		gatewaysWithState[i].Active = gateway.Name == activeGatewayName

		if status, found := gatewayStatusMap[gateway.StatusName]; found {
			gatewaysWithState[i].HasKnownStatus = true
			gatewaysWithState[i].RoundtripTime = status.RoundtripTime()
			gatewaysWithState[i].Online = status.Online()
		}
	}

	w.Header().Add("Content-Type", "text/html")

	err = templateViewGateways.Execute(w, struct {
		Gateways []struct {
			Name   string
			Label  string
			Active bool

			HasKnownStatus bool
			RoundtripTime  string
			Online         bool
		}
		Source string
	}{
		Gateways: gatewaysWithState,
		Source:   ip,
	})
	if err != nil {
		log.Println("error rendering ViewGateways:", err)
	}
}

func handlerSetGateway(w http.ResponseWriter, r *http.Request) {
	gateways := cfg.Gateways

	userSubmittedGateway := r.FormValue("gateway")
	var gateway gateway

	for _, gateway = range gateways {
		if gateway.Name == userSubmittedGateway {
			break
		}
	}

	if gateway.Name == "" {
		gateway.Name = deleteDork
		gateway.Label = deleteDork
	}

	ip, err := getSource(r)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("could not get source"))
		return
	}

	_, err = setGateway(cfg.RemoteInterface, ip, gateway.Name, gateway.Label)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("failed to set gateway"))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func bootServer(ctx context.Context) chan error {
	router := makeRouter([]routeDef{
		routeDef{"GET", "/", "ViewGateways", handlerViewGateways},
		routeDef{"POST", "/", "SetGateway", handlerSetGateway},
	})

	src := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	errorChannel := make(chan error, 1)

	go func() {
		log.Println("listening on", src.Addr)
		errorChannel <- src.ListenAndServe()
	}()

	go func() {
		select {
		case <-ctx.Done():
			gracefulCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			src.Shutdown(gracefulCtx)
			<-errorChannel
		}
	}()

	return errorChannel
}
