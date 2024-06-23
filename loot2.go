package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
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
				Ratio:        1 / float64(len(datos_binance)), // Ratio fijo, puedes ajustarlo según tu necesidad
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

	// Seleccionar 10 monedas aleatorias sin repetición
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	monedasAleatorias := make([]Monedas, 0, 10)
	monedasSeleccionadas := make(map[int]bool)

	for len(monedasAleatorias) < 10 {
		indice := rng.Intn(len(monedas))
		if !monedasSeleccionadas[indice] {
			monedasAleatorias = append(monedasAleatorias, monedas[indice])
			monedasSeleccionadas[indice] = true
		}
	}

	// Ajustar probabilidades para las 10 monedas seleccionadas
	sumaProbabilidades := 0.0
	for _, moneda := range monedasAleatorias {
		sumaProbabilidades += moneda.Probabilidad
	}
	for i := range monedasAleatorias {
		monedasAleatorias[i].Probabilidad /= sumaProbabilidades
	}

	// Retornar las monedas y ningún error
	return monedasAleatorias, nil
}

func sumaProbabilidadesAproximada(monedas []Monedas) bool {
	sumaProbabilidad := 0.0
	for _, moneda := range monedas {
		sumaProbabilidad += moneda.Probabilidad
	}

	return math.Abs(sumaProbabilidad-1.0) < 0.0001
}

func main() {
	monedas, err := crear_monedas()
	if err != nil {
		fmt.Printf("Error al obtener las monedas de Binance: %v\n", err)
		return
	}

	for _, moneda := range monedas {
		fmt.Printf("Nombre: %s\n", moneda.Nombre)
		fmt.Printf("Siglas: %s\n", moneda.Siglas)
		fmt.Printf("Icono: %s\n", moneda.Icono)
		fmt.Printf("Ganancia: %.8f%%\n", moneda.Ganancia*100)
		fmt.Printf("Probabilidad: %.8f%%\n", moneda.Probabilidad*100)
		fmt.Printf("Ratio: %.8f\n", moneda.Ratio)
		fmt.Println("---")
	}

	if sumaProbabilidadesAproximada(monedas) {
		fmt.Println("La suma de las probabilidades es aproximadamente igual a 1.")
	} else {
		fmt.Println("Error: La suma de las probabilidades no es aproximadamente igual a 1.")
	}
}
