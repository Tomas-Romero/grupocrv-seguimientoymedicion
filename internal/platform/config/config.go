// Package config lee la configuracion del entorno y la valida.
//
// La funcion Cargar recibe el lector de variables como parametro en lugar de
// llamar a os.Getenv directamente. Eso es lo que permite testearla sin ensuciar
// el entorno del proceso de tests.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Errores del paquete. Se exponen como valores para poder compararlos con
// errors.Is en los tests y en quien llame.
var (
	ErrFaltaDatabaseURL           = errors.New("falta la variable DATABASE_URL")
	ErrAddrInvalida               = errors.New("HTTP_ADDR debe tener el formato :puerto o host:puerto")
	ErrCargarDatosEjemploInvalido = errors.New("CARGAR_DATOS_EJEMPLO debe ser true o false")
)

// Config es la configuracion de la aplicacion.
type Config struct {
	DatabaseURL string
	HTTPAddr    string
	LogLevel    string

	// CargarDatosEjemplo pide cargar los datos de ejemplo al arrancar (DEF-002).
	// Solo lo activa el perfil full del compose.
	CargarDatosEjemplo bool
}

// LectorEntorno devuelve el valor de una variable de entorno. os.Getenv lo cumple.
type LectorEntorno func(clave string) string

// Cargar arma la configuracion a partir del entorno y la valida.
// Aplica valores por defecto para todo lo que no sea obligatorio.
func Cargar(getenv LectorEntorno) (Config, error) {
	cfg := Config{
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
		HTTPAddr:    valorOPorDefecto(getenv("HTTP_ADDR"), ":8080"),
		LogLevel:    valorOPorDefecto(getenv("LOG_LEVEL"), "info"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, ErrFaltaDatabaseURL
	}
	if !strings.Contains(cfg.HTTPAddr, ":") {
		return Config{}, fmt.Errorf("%w: recibido %q", ErrAddrInvalida, cfg.HTTPAddr)
	}

	cargarDatos, err := leerBooleano(getenv("CARGAR_DATOS_EJEMPLO"))
	if err != nil {
		return Config{}, fmt.Errorf("%w: recibido %q", ErrCargarDatosEjemploInvalido, getenv("CARGAR_DATOS_EJEMPLO"))
	}
	cfg.CargarDatosEjemplo = cargarDatos

	return cfg, nil
}

// leerBooleano interpreta una variable booleana. Sin definir (o solo con
// espacios) es falsa; si no, acepta lo mismo que strconv.ParseBool.
func leerBooleano(valor string) (bool, error) {
	v := strings.TrimSpace(valor)
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}

func valorOPorDefecto(valor, porDefecto string) string {
	if v := strings.TrimSpace(valor); v != "" {
		return v
	}
	return porDefecto
}
