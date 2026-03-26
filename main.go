package main

import (
	"crypto/tls"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"
)

//go:embed tpl.html
var htmlTemplate string

var nodeName string
var httpsPort int

// Structure de données pour passer l'IP au template.
type PageData struct {
	IP            string
	LocalIP       string
	NodeName      string
	Proto         string
	Headers       http.Header
	TLSVersion    string
	TLSServerName string
	ALPN          string
}

// ipHandler est le gestionnaire pour nos requêtes HTTP.
func ipHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.RemoteAddr contient l'adresse IP et le port source (ex: "192.0.2.1:12345" ou "[2001:db8::1]:54321").
		// net.SplitHostPort sépare correctement l'hôte (IP) du port pour IPv4 et IPv6.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Si SplitHostPort échoue, c'est peut-être que le port n'est pas présent.
			// Dans ce cas, RemoteAddr est probablement juste l'IP.
			// C'est moins courant mais on le gère pour plus de robustesse.
			log.Printf("Impossible de séparer l'hôte et le port pour %q: %v. Utilisation de la valeur brute.", r.RemoteAddr, err)
			ip = r.RemoteAddr
		}
		proto := r.Proto
		if proto != "HTTP/3.0" {
			w.Header().Add("Alt-Svc", fmt.Sprintf(`h3=":%d"; ma=900`, httpsPort))
		}

		// On prépare les données pour le template.
		data := PageData{
			IP:       ip,
			NodeName: nodeName,
			Proto:    proto,
			Headers:  r.Header,
		}
		if r.TLS != nil {
			data.TLSVersion = tls.VersionName(r.TLS.Version)
			data.ALPN = r.TLS.NegotiatedProtocol
			data.TLSServerName = r.TLS.ServerName
		}

		// On exécute le template en lui passant les données.
		// Le résultat est écrit dans http.ResponseWriter.
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Erreur lors de l'exécution du template: %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}
}

func main() {
	// Définition des flags pour la configuration de l'adresse et des ports.
	// L'adresse vide "" ou "[::]" signifie une écoute sur toutes les interfaces réseau (IPv4 et IPv6).
	addr := flag.String("addr", "", "Adresse d'écoute (ex: 127.0.0.1, [::1]). Laisser vide pour toutes les interfaces.")
	httpPort := flag.Int("http-port", 80, "Port d'écoute pour HTTP")
	_httpsPort := flag.Int("https-port", 443, "Port d'écoute pour HTTPS")
	certPath := flag.String("cert", "", "Path to the certificate file for HTTPS")
	keyPath := flag.String("key", "", "Path to the private key file for HTTPS")
	tlsEnabled := flag.Bool("tls", false, "Enable HTTPS with TLS")
	tpl := flag.String("tpl", "", "Use another template")
	flag.Parse()

	httpsPort = *_httpsPort

	// Verify that TLS is enabled if certificate files are specified
	if *certPath != "" || *keyPath != "" {
		if !*tlsEnabled {
			log.Fatal("TLS certificate files specified but --tls flag not set. Use --tls to enable HTTPS.")
		}
	}

	nodeName = os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName, _ = os.Hostname()
	}
	if *tpl != "" {
		f, err := os.Open(*tpl)
		if err == nil {
			buf, err := io.ReadAll(f)
			if err == nil {
				htmlTemplate = string(buf)
			}
		}
	}

	// Compilation du template HTML une seule fois au démarrage pour de meilleures performances.
	tmpl, err := template.New("ipPage").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Erreur: Impossible de compiler le template HTML. %v", err)
	}

	// Création des adresses d'écoute pour HTTP, HTTPS et QUIC (HTTP/3)
	httpListenAddr := fmt.Sprintf("%s:%d", *addr, *httpPort)
	httpsListenAddr := fmt.Sprintf("%s:%d", *addr, httpsPort)

	// Configuration TLS pour HTTPS
	if *tlsEnabled && *certPath != "" && *keyPath != "" {
		// Load certificate and key
		_, err := tls.LoadX509KeyPair(*certPath, *keyPath)
		if err != nil {
			log.Fatalf("Erreur lors du chargement du certificat: %v", err)
		}
	}

	mux := http.NewServeMux()

	// Enregistrement de notre gestionnaire pour la racine du site "/".
	mux.HandleFunc("/", ipHandler(tmpl))
	mux.HandleFunc("/proto", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Proto)
	})

	// Démarrage du serveur HTTP/3
	if *tlsEnabled {
		go func() {
			log.Printf("Démarrage HTTP/3 sur %s", httpListenAddr)
			if err := http3.ListenAndServeTLS(httpsListenAddr, *certPath, *keyPath, mux); err != nil {
				log.Fatalf("Erreur: Impossible de démarrer le serveur HTTP/3. %v", err)
			}
		}()
	}

	// Start HTTP server
	log.Printf("Démarrage du servuer HTTP/1.1 sur %s", httpListenAddr)
	if err := http.ListenAndServe(httpListenAddr, mux); err != nil {
		log.Fatalf("Erreur: Impossible de démarrer le serveur HTTP. %v", err)
	}

}
