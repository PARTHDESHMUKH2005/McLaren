package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type car struct {
	Name     string `json:"name"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Desc     string `json:"desc"`
	Price    string `json:"price"`
	ImageURL string `json:"image_url"`
}

type specifications struct {
	Engine       string `json:"engine"`
	Horsepower   int    `json:"horsepower"`
	Torque       string `json:"torque"`
	TopSpeed     string `json:"topspeed"`
	Acceleration string `json:"acceleration"`
	Weight       string `json:"weight"`
	Transmission string `json:"transmission"`
}

type feature struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Icon  string `json:"icon"`
}

func get_carinfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only get request allowed", http.StatusMethodNotAllowed)
		return
	}

	Car := car{
		Name:     "McLaren 765LT",
		Model:    "765LT",
		Year:     2021,
		Desc:     "The 765LT is McLaren's most powerful and track-focused LT model, delivering extreme performance with reduced weight.",
		Price:    "$338,500",
		ImageURL: "https://media.gq-magazine.co.uk/photos/5e5d438f43621e0008b6c0ca/16:9/w_1280,c_limit/20200302-765-09.jpg",
	}
	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(Car)

}
func getSpecifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	specs := specifications{
		Engine:       "4.0L Twin-Turbo V8",
		Horsepower:   765,
		Torque:       "590 lb-ft",
		TopSpeed:     "205 mph",
		Acceleration: "0-60 mph in 2.7 seconds",
		Weight:       "2,952 lbs",
		Transmission: "7-Speed Dual-Clutch",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(specs)
}
func getFeatures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	Features := []feature{
		{
			ID:    1,
			Title: "Longtail Aerodynamics",
			Desc:  "Extended rear end with active aerodynamics for maximum downforce and stability",
			Icon:  "🏎️",
		},
		{
			ID:    2,
			Title: "Carbon Fiber Body",
			Desc:  "Lightweight carbon fiber construction reduces weight by 80kg compared to 720S",
			Icon:  "⚡",
		},
		{
			ID:    3,
			Title: "Track Telemetry",
			Desc:  "Advanced telemetry system to record and analyze lap times and performance data",
			Icon:  "📊",
		},
		{
			ID:    4,
			Title: "Titanium Exhaust",
			Desc:  "Ultra-lightweight titanium exhaust system with distinctive LT soundtrack",
			Icon:  "🔊",
		},
		{
			ID:    5,
			Title: "Racing Seats",
			Desc:  "Carbon fiber racing seats with Alcantara trim for ultimate support",
			Icon:  "💺",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Features)
}
func homehandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/main1.html")
}

func main() {
	http.HandleFunc("/", homehandler)
	http.HandleFunc("/api/car", get_carinfo)
	http.HandleFunc("/api/info", getSpecifications)
	http.HandleFunc("/api/features", getFeatures)

	fmt.Println("mclaren server running")
	fmt.Println("server running on http://localhost:5001")
	err := http.ListenAndServe(":5001", nil)
	if err != nil {
		log.Fatal("server failed to start", err)
	}
}
