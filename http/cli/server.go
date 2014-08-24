package cli

import (
	httputil ".."

	"flag"
	"net/http"
)

var (
	listenAddress  = flag.String("listen", "localhost:8080", "The address to listen on.")
	wrapLogHandler = flag.Bool("log-requests", false, "Should requests be logged to stdout")

	httpsUse             = flag.Bool("https-enable", false, "Enable HTTPS listening in favor of HTTP.")
	httpsCertificateFile = flag.String("https-certificate", "server.cert", "The certificate to use for SSL.")
	httpsKeyFile         = flag.String("https-key", "server.key", "The keyfile to use for SSL.")
)

func StartHttpInterface(handler http.Handler) {
	if *wrapLogHandler {
		handler = &httputil.RequestLogger{handler}
	}

	if *httpsUse {
		if err := http.ListenAndServeTLS(*listenAddress, *httpsCertificateFile, *httpsKeyFile, handler); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(*listenAddress, handler); err != nil {
			panic(err)
		}
	}
}
