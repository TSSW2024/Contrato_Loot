package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/cors"
)

type respuesta_binance struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Monedas struct {
	Icono        string
	Nombre       string
	Siglas       string
	Ratio        float64
	Ganancia     float64
	Probabilidad float64
}

type coingeckocoin struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Image  string `json:"image"`
}

func obtener_datos_geckocoin() (map[string]coingeckocoin, error) {
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}
	var coins []coingeckocoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON de CoinGecko: %v", err)
	}
	coinMap := make(map[string]coingeckocoin)
	for _, coin := range coins {
		coinMap[strings.ToUpper(coin.Symbol)] = coin
	}
	return coinMap, nil
}

func crear_monedas() ([]Monedas, error) {
	url := "https://api.binance.com/api/v3/ticker/price"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}

	var datos_binance []respuesta_binance
	if err := json.NewDecoder(resp.Body).Decode(&datos_binance); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON: %v", err)
	}

	coinGeckoMap, err := obtener_datos_geckocoin()
	if err != nil {
		return nil, err
	}

	var monedas []Monedas
	for _, datos := range datos_binance {
		if strings.HasSuffix(datos.Symbol, "USDT") {
			precio, err := strconv.ParseFloat(datos.Price, 64)
			if err != nil {
				return nil, fmt.Errorf("error al convertir el precio a float: %v", err)
			}
			siglas := strings.TrimSuffix(datos.Symbol, "USDT")
			coinData, exists := coinGeckoMap[siglas]
			if !exists {
				continue
			}

			ganancia := 1 / precio // Calcular ganancia como inversa del precio
			moneda := Monedas{
				Icono:        coinData.Image,
				Nombre:       coinData.Name,
				Siglas:       datos.Symbol,
				Ratio:        1 / float64(len(datos_binance)), // Ratio fijo
				Ganancia:     ganancia,
				Probabilidad: ganancia, // Probabilidad proporcional a la ganancia
			}
			monedas = append(monedas, moneda)
		}
	}

	// Normalizar las probabilidades proporcionalmente a la ganancia
	sumaGanancias := 0.0
	for _, moneda := range monedas {
		sumaGanancias += moneda.Ganancia
	}
	for i := range monedas {
		monedas[i].Probabilidad /= sumaGanancias
	}

	// Retornar todas las monedas y ningún error
	return monedas, nil
}

func crear_caja_1(monedas []Monedas) []Monedas {
	var caja []Monedas
	for _, moneda := range monedas {
		if moneda.Ganancia*100 < 1.1 {
			caja = append(caja, moneda)
		}
	}
	if len(caja) > 10 {
		caja = caja[:10]
	}
	return caja
}
func crear_caja_2(monedas []Monedas) []Monedas {
	var caja []Monedas
	var valorAlto Monedas
	for _, moneda := range monedas {
		if moneda.Ganancia*100 >= 80 {
			caja = append(caja, moneda)
		} else if moneda.Ganancia*100 < 0.5 {
			valorAlto = moneda
		}
	}
	if len(caja) > 9 {
		caja = caja[:9]
	}
	if valorAlto != (Monedas{}) {
		caja = append(caja, valorAlto)
	}
	return caja
}
func ajustar_probabilidades_caja(caja []Monedas) {
	sumaGanancias := 0.0
	for _, moneda := range caja {
		sumaGanancias += moneda.Ganancia
	}
	for i := range caja {
		caja[i].Probabilidad = caja[i].Ganancia / sumaGanancias
	}
}

func handlerCaja1(w http.ResponseWriter, r *http.Request) {
	monedas, err := crear_monedas()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener las monedas de Binance: %v", err), http.StatusInternalServerError)
		return
	}

	caja1 := crear_caja_1(monedas)
	ajustar_probabilidades_caja(caja1)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(caja1); err != nil {
		http.Error(w, fmt.Sprintf("Error al codificar las monedas de la caja 1 a JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func handlerCaja2(w http.ResponseWriter, r *http.Request) {
	monedas, err := crear_monedas()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener las monedas de Binance: %v", err), http.StatusInternalServerError)
		return
	}

	caja2 := crear_caja_2(monedas)
	ajustar_probabilidades_caja(caja2)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(caja2); err != nil {
		http.Error(w, fmt.Sprintf("Error al codificar las monedas de la caja 2 a JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	// Define tus manejadores
	http.HandleFunc("/caja1", handlerCaja1)
	http.HandleFunc("/caja2", handlerCaja2)

	// Configura CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Permite todos los orígenes, puedes cambiar esto por los orígenes específicos que necesites
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Envuelve tu manejador HTTP con el manejador CORS
	handler := c.Handler(http.DefaultServeMux)

	port := "8082"
	fmt.Printf("Servidor escuchando en el puerto %s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Printf("Error al iniciar el servidor: %v\n", err)
	}
}
