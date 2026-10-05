// Package migraciones guarda las migraciones SQL de la base embebidas en el
// binario, para que la aplicacion se migre sola al arrancar (DEF-002).
//
// La CLI de goose (`make migrate`) lee los mismos archivos desde el disco. Este
// archivo no le molesta: goose ignora los .go que no empiezan con un numero de
// version.
package migraciones

import "embed"

// Archivos son las migraciones numeradas de esta carpeta (NNNNN_descripcion.sql).
//
//go:embed *.sql
var Archivos embed.FS
