package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

//MAGAZYN:
// 100 zniczy [znicz1..znicz100]
// 50 wiązanek [wiązanka1..wiązanka50]

// 2 babki pobierają znicze 2 wiązanki
// kazda jednorazowo pobiera 1 sztukę

// kosz na znicze 10
// kosz na wiązanki 10

// 5 posłańcow
// maksymalnie 1 wiazanka 2 znicze

var znicze_start = 100
var wiazanki_start = 50

var znicze = znicze_start
var wiazanki = wiazanki_start

var kosz_znicze = 0
var kosz_wiazanki = 0

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/magazyn", magazyn)
	mux.HandleFunc("/kosz", kosz)
	mux.HandleFunc("/znicze/", Znicze)
	mux.HandleFunc("/wiazanki/", Wiazanki)
	mux.HandleFunc("/poslaniec/", poslaniec)
	log.Fatal(http.ListenAndServe(":3000", recoverMw(mux, true)))
}

func recoverMw(app http.Handler, dev bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				if !dev {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, string(stack))
			}
		}()

		nw := &responseWriter{ResponseWriter: w}
		app.ServeHTTP(nw, r)
		nw.flush()
	}
}

type responseWriter struct {
	http.ResponseWriter
	writes [][]byte
	status int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.writes = append(rw.writes, b)
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter does not support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func (rw *responseWriter) flush() error {
	if rw.status != 0 {
		rw.ResponseWriter.WriteHeader(rw.status)
	}
	for _, write := range rw.writes {
		_, err := rw.ResponseWriter.Write(write)
		if err != nil {
			return err
		}
	}
	return nil
}

func magazyn(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w,
		"W magazynie:",
		"\nZnicze: ",
		znicze,
		"\nWiazanki: ",
		wiazanki,
	)
}
func kosz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w,
		"W koszu:",
		"\nZnicze: ",
		kosz_znicze,
		"\nWiazanki: ",
		kosz_wiazanki,
	)
}

func Znicze(w http.ResponseWriter, r *http.Request) {
	babka_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/znicze/"))

	if err != nil {
		panic("Nie mozna pobrac numeru babki!")
	}
	if babka_numer > 2 {
		msg := fmt.Sprintf("Maksymalna ilosc babek od zniczy to 2")
		panic(msg)
	}
	if znicze == 0 {
		panic("Brak zniczy na magazynie")
	}
	if kosz_znicze == 10 {
		msg := fmt.Sprintf("Kosz na znicze pelny, maksymalna ilosc to 10")
		panic(msg)
	}

	kosz_znicze = kosz_znicze + 1
	znicze = znicze - 1

	fmt.Fprintln(w, "Babka", babka_numer, "pobrala znicz ", math.Abs(float64(znicze-znicze_start)), " z magazynu do kosza.")
}

func Wiazanki(w http.ResponseWriter, r *http.Request) {
	babka_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/wiazanki/"))

	if err != nil {
		panic("Nie mozna pobrac numeru babki!")
	}
	if babka_numer > 2 {
		msg := fmt.Sprintf("Maksymalna ilosc babek od wiazanek to 2")
		panic(msg)
	}
	if wiazanki == 0 {
		panic("Brak wiazanek na magazynie")
	}
	if kosz_wiazanki == 10 {
		msg := fmt.Sprintf("Kosz na wiazanki jest pelny, maksymalna ilosc to 10")
		panic(msg)
	}

	kosz_wiazanki = kosz_wiazanki + 1
	wiazanki = wiazanki - 1

	fmt.Fprintln(w, "Babka", babka_numer, "pobrala wiazanke ", math.Abs(float64(wiazanki-wiazanki_start)), " z magazynu do kosza.")
}

func poslaniec(w http.ResponseWriter, r *http.Request) {
	poslaniec_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/poslaniec/"))

	if err != nil {
		panic("Nie mozna pobrac numeru babki!")
	}
	if poslaniec_numer > 5 {
		msg := fmt.Sprintf("Maksymalna ilosc poslancow to 5")
		panic(msg)
	}
	if kosz_wiazanki == 0 && kosz_znicze == 0 {
		panic("Kosz jest pusty!")
	}
	msg := fmt.Sprintf("")
	if kosz_znicze >= 2 {
		kosz_znicze = kosz_znicze - 2
		msg = msg + fmt.Sprintf("Poslaniec %d pobiera %d znicze", poslaniec_numer, 2)
	} else {
		msg = msg + fmt.Sprintf("Poslaniec %d pobiera %d znicze", poslaniec_numer, kosz_znicze)
		kosz_znicze = 0
	}

	if kosz_wiazanki >= 1 {
		kosz_wiazanki = kosz_wiazanki - 1
		msg = msg + fmt.Sprintf("\nPoslaniec %d pobiera %d wiazanki", poslaniec_numer, 1)
	} else {
		msg = msg + fmt.Sprintf("\nPoslaniec %d pobiera %d wiazanki", poslaniec_numer, kosz_wiazanki)
		kosz_wiazanki = 0
	}

	fmt.Fprintln(w, msg)
}
