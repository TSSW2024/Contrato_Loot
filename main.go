package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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

type coinapicoin struct {
	AssetID string `json:"asset_id"`
	Name    string `json:"name"`
	Url     string `json:"url"`
}

func obtener_datos_coinapi_icons(apiKey string) (map[string]string, error) {
	url := "https://rest.coinapi.io/v1/assets/icons/55"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}
	req.Header.Set("X-CoinAPI-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}

	var coins []coinapicoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON de CoinAPI: %v", err)
	}

	iconMap := make(map[string]string)
	for _, coin := range coins {
		iconMap[strings.ToUpper(coin.AssetID)] = coin.Url
	}
	return iconMap, nil
}

func obtener_datos_coinapi_assets(apiKey string) (map[string]string, error) {
	url := "https://rest.coinapi.io/v1/assets"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}
	req.Header.Set("X-CoinAPI-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: respuesta HTTP %v", resp.StatusCode)
	}

	var coins []coinapicoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta JSON de CoinAPI: %v", err)
	}

	nameMap := make(map[string]string)
	for _, coin := range coins {
		nameMap[strings.ToUpper(coin.AssetID)] = coin.Name
	}
	return nameMap, nil
}

func obtener_datos_coinapi(apiKey string) (map[string]coinapicoin, error) {
	icons, err := obtener_datos_coinapi_icons(apiKey)
	if err != nil {
		return nil, err
	}

	names, err := obtener_datos_coinapi_assets(apiKey)
	if err != nil {
		return nil, err
	}

	coinMap := make(map[string]coinapicoin)
	for assetID, icon := range icons {
		name, exists := names[assetID]
		if exists {
			coinMap[assetID] = coinapicoin{
				AssetID: assetID,
				Name:    name,
				Url:     icon,
			}
		}
	}
	return coinMap, nil
}

func crear_monedas(apiKey string) ([]Monedas, error) {
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

	coinAPIMap, err := obtener_datos_coinapi(apiKey)
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
			coinData, exists := coinAPIMap[siglas]
			if !exists {
				continue
			}

			ganancia := 1 / precio
			moneda := Monedas{
				Icono:        coinData.Url,
				Nombre:       coinData.Name,
				Siglas:       datos.Symbol,
				Ratio:        1 / float64(len(datos_binance)),
				Ganancia:     ganancia,
				Probabilidad: ganancia,
			}
			monedas = append(monedas, moneda)
		}
	}

	// Normalizar las probabilidades
	sumaGanancias := 0.0
	for _, moneda := range monedas {
		sumaGanancias += moneda.Ganancia
	}
	for i := range monedas {
		monedas[i].Probabilidad = monedas[i].Ganancia / sumaGanancias
	}

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
func handler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("COINAPI_KEY")
	if apiKey == "" {
		http.Error(w, "Error: API key de CoinAPI no configurada", http.StatusInternalServerError)
		return
	}

	monedas, err := crear_monedas(apiKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener las monedas: %v", err), http.StatusInternalServerError)
		return
	}

	caja1 := crear_caja_1(monedas)
	caja2 := crear_caja_2(monedas)
	ajustar_probabilidades_caja(caja1)
	ajustar_probabilidades_caja(caja2)

	for _, moneda := range caja1 {
		fmt.Fprintf(w, "Caja 1 - Nombre: %s\n", moneda.Nombre)
		fmt.Fprintf(w, "Siglas: %s\n", moneda.Siglas)
		fmt.Fprintf(w, "Icono: %s\n", moneda.Icono)
		fmt.Fprintf(w, "Ganancia: %.8f%%\n", moneda.Ganancia*100)
		fmt.Fprintf(w, "Probabilidad: %.8f%%\n", moneda.Probabilidad*100)
		fmt.Fprintf(w, "Ratio: %.8f\n", moneda.Ratio)
		fmt.Fprintln(w, "---")
	}

	for _, moneda := range caja2 {
		fmt.Fprintf(w, "Caja 2 - Nombre: %s\n", moneda.Nombre)
		fmt.Fprintf(w, "Siglas: %s\n", moneda.Siglas)
		fmt.Fprintf(w, "Icono: %s\n", moneda.Icono)
		fmt.Fprintf(w, "Ganancia: %.8f%%\n", moneda.Ganancia*100)
		fmt.Fprintf(w, "Probabilidad: %.8f%%\n", moneda.Probabilidad*100)
		fmt.Fprintf(w, "Ratio: %.8f\n", moneda.Ratio)
		fmt.Fprintln(w, "---")
	}
}

func main() {
	// Cargar las variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error al cargar el archivo .env: %v\n", err)
		return
	}

	http.HandleFunc("/", handler)
	port := "8082"
	fmt.Printf("Servidor escuchando en el puerto %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error al iniciar el servidor: %v\n", err)
	}
}
