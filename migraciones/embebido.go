// Package migraciones guarda las migraciones SQL de la base y los datos de
// ejemplo embebidos en el binario, para que la aplicacion se migre sola al
// arrancar y pueda cargar los datos en el perfil full (DEF-002).
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

// DatosEjemplo es el SQL de los datos de ejemplo. Esta en una subcarpeta para que
// ni la CLI de goose ni la app lo tomen como migracion: las dos buscan *.sql solo
// en esta carpeta, sin entrar en subcarpetas.
//
//go:embed datos/ejemplo.sql
var DatosEjemplo string
