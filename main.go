package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

// Constante contenant notre template HTML.
// L'utilisation de ` backticks ` permet d'écrire sur plusieurs lignes.
//
//go:embed tpl.html
var htmlTemplate string

// Structure de données pour passer l'IP au template.
type PageData struct {
	IP      string
	LocalIP string
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

		// On prépare les données pour le template.
		data := PageData{IP: ip}

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
	// Définition des flags pour la configuration de l'adresse et du port.
	// L'adresse vide "" ou "[::]" signifie une écoute sur toutes les interfaces réseau (IPv4 et IPv6).
	addr := flag.String("addr", "", "Adresse d'écoute (ex: 127.0.0.1, [::1]). Laisser vide pour toutes les interfaces.")
	port := flag.Int("port", 8080, "Port d'écoute")
	flag.Parse()

	// Compilation du template HTML une seule fois au démarrage pour de meilleures performances.
	tmpl, err := template.New("ipPage").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Erreur: Impossible de compiler le template HTML. %v", err)
	}

	// Création de l'adresse d'écoute complète.
	listenAddr := fmt.Sprintf("%s:%d", *addr, *port)

	// Enregistrement de notre gestionnaire pour la racine du site "/".
	http.HandleFunc("/", ipHandler(tmpl))

	// Affichage d'un message de démarrage dans la console.
	log.Printf("Serveur web démarré. Écoute sur: http://localhost:%d (et sur %s)", *port, listenAddr)

	// Démarrage du serveur HTTP. log.Fatal s'exécutera en cas d'erreur au démarrage.
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Erreur: Impossible de démarrer le serveur. %v", err)
	}
}
