package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type respuestaBinance struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Moneda struct {
	Icono        string
	Nombre       string
	Siglas       string
	Ratio        float64
	Ganancia     float64
	Probabilidad float64
	ID           string // Nuevo campo para almacenar el ID de CoinGecko
}

type coingeckoCoin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Image  struct {
		Thumb string `json:"thumb"`
	} `json:"image"`
}

func obtenerDatosCoinGecko(symbol string) (coingeckoCoin, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return coingeckoCoin{}, fmt.Errorf("error al hacer la solicitud HTTP a CoinGecko: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return coingeckoCoin{}, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}

	var coin coingeckoCoin
	if err := json.NewDecoder(resp.Body).Decode(&coin); err != nil {
		return coingeckoCoin{}, fmt.Errorf("error al decodificar la respuesta JSON de CoinGecko: %v", err)
	}

	return coin, nil
}

func obtenerDatosCoinGeckoLista() (map[string]coingeckoCoin, error) {
	url := "https://api.coingecko.com/api/v3/coins/list"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}
	var coins []coingeckoCoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON de CoinGecko: %v", err)
	}
	coinMap := make(map[string]coingeckoCoin)
	for _, coin := range coins {
		coinMap[strings.ToUpper(coin.Symbol)] = coin
	}
	return coinMap, nil
}

func crearMonedas() ([]Moneda, error) {
	// Obtener datos de Binance
	datosBinance, err := obtenerDatosBinance()
	if err != nil {
		return nil, err
	}

	// Obtener datos de CoinGecko
	coinGeckoMap, err := obtenerDatosCoinGeckoLista()
	if err != nil {
		return nil, err
	}

	// Contar el número total de monedas USDT para calcular probabilidades
	totalUSDT := 0.0
	for _, datos := range datosBinance {
		if strings.HasSuffix(datos.Symbol, "USDT") {
			totalUSDT++
		}
	}

	// Slice para almacenar las monedas a retornar
	var monedas []Moneda

	// Iterar sobre los datos de Binance
	for _, datos := range datosBinance {
		if strings.HasSuffix(datos.Symbol, "USDT") {
			// Convertir el precio a float
			precio, err := strconv.ParseFloat(datos.Price, 64)
			if err != nil {
				return nil, fmt.Errorf("error al convertir el precio a float: %v", err)
			}

			// Obtener las siglas de la moneda
			siglas := strings.TrimSuffix(datos.Symbol, "USDT")

			// Verificar si la moneda existe en CoinGecko
			coinData, exists := coinGeckoMap[siglas]
			if !exists {
				continue
			}

			// Calcular la probabilidad
			probabilidad := 1 / totalUSDT

			// Obtener la URL del ícono desde CoinGecko
			iconoURL := coinData.Image.Thumb

			// Crear la estructura Moneda
			moneda := Moneda{
				Icono:        iconoURL,
				Nombre:       coinData.Name,
				Siglas:       datos.Symbol,
				Ratio:        1 / totalUSDT,
				Ganancia:     1 / precio,
				Probabilidad: probabilidad,
				ID:           coinData.ID,
			}

			// Agregar la moneda al slice de monedas
			monedas = append(monedas, moneda)
		}
	}

	// Ajustar probabilidades para asegurar que sumen aproximadamente 1
	sumaProbabilidades := 0.0
	for _, moneda := range monedas {
		sumaProbabilidades += moneda.Probabilidad
	}
	for i := range monedas {
		monedas[i].Probabilidad /= sumaProbabilidades
	}

	// Retornar las monedas y ningún error
	return monedas, nil
}

func sumaProbabilidadesAproximada(monedas []Moneda) bool {
	sumaProbabilidad := 0.0
	for _, moneda := range monedas {
		sumaProbabilidad += moneda.Probabilidad
	}
	return math.Abs(sumaProbabilidad-1.0) < 0.0001
}

func main() {
	monedas, err := crearMonedas()
	if err != nil {
		fmt.Printf("Error al obtener las monedas: %v\n", err)
		return
	}

	for _, moneda := range monedas {
		fmt.Printf("Nombre: %s\n", moneda.Nombre)
		fmt.Printf("Siglas: %s\n", moneda.Siglas)
		fmt.Printf("Icono: %s\n", moneda.Icono)
		fmt.Printf("Ganancia: %.8f%%\n", moneda.Ganancia*100)
		fmt.Printf("Probabilidad: %.8f\n", moneda.Probabilidad)
		fmt.Printf("Ratio: %.8f\n", moneda.Ratio)
		fmt.Println("---")
	}

	if sumaProbabilidadesAproximada(monedas) {
		fmt.Println("La suma de las probabilidades es aproximadamente igual a 1.")
	} else {
		fmt.Println("Error: La suma de las probabilidades no es aproximadamente igual a 1.")
	}
}

func obtenerDatosBinance() ([]respuestaBinance, error) {
	url := "https://api.binance.com/api/v3/ticker/price"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP a Binance: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}

	var datos []respuestaBinance
	if err := json.NewDecoder(resp.Body).Decode(&datos); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON de Binance: %v", err)
	}

	return datos, nil
}
