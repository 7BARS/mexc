package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type TradeResponse struct {
	C string `json:"c,omitempty"`
	D D      `json:"d,omitempty"`
	S string `json:"s,omitempty"`
	T int64  `json:"t,omitempty"`
}
type Deals struct {
	P string `json:"p,omitempty"`
	V string `json:"v,omitempty"`
	S int    `json:"S,omitempty"`
	T int64  `json:"t,omitempty"`
}
type D struct {
	Deals []Deals `json:"deals,omitempty"`
	E     string  `json:"e,omitempty"`
}

func main() {
	db, err := newPostgresql()
	if err != nil {
		log.Fatal(err)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "wbs.mexc.com", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
				return
			}
			var orders TradeResponse
			err = json.Unmarshal(message, &orders)
			if err != nil {
				log.Fatal("unmarshal error:", err)
			}

			deals := make([]deal, 0, len(orders.D.Deals))
			for _, v := range orders.D.Deals {
				priceType := "buy"
				if v.T == 2 {
					priceType = "sell"
				}
				price, err := strconv.ParseFloat(v.P, 64)
				if err != nil {
					log.Fatal(err)
				}
				volume, err := strconv.ParseFloat(v.V, 64)
				if err != nil {
					log.Fatalln(err)
				}
				deals = append(deals, deal{
					Price:  price,
					Type:   int64(v.S),
					Time:   v.T,
					Volume: volume,
				})
				log.Printf("time: %v, price: %v, volume: %v, type: %v", v.T, v.P, v.V, priceType)
			}
			if len(deals) == 0 {
				log.Println(string(message))
				continue
			}
			err = db.Create(deals)
		}
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	err = c.WriteMessage(websocket.TextMessage, []byte("{ \"method\":\"SUBSCRIPTION\", \"params\":[\"spot@public.deals.v3.api@GRIMACEUSDT\"] }"))
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"PING\"}"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
