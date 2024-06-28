# Usa una imagen de Golang como base
FROM golang:latest

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copia el código actual al contenedor
COPY . .

# Descarga las dependencias del proyecto
RUN go mod download

# Compila el código Go dentro del contenedor
RUN go build -o main .

# Expone el puerto 8082 en el contenedor
EXPOSE 8082

# Comando por defecto para ejecutar la aplicación
CMD ["./main"]
