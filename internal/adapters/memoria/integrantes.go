package memoria

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// Base cumple el puerto del caso de uso de US-003.
var _ app.RepositorioRegistrarIntegrante = (*Base)(nil)

// RegistrarIntegrante guarda al integrante con un ID nuevo, como el RETURNING de
// Postgres (RN-003-10). Replica lo que hace la persistencia (US-003):
//
//   - si el proyecto no existe, devuelve app.ErrProyectoInexistente;
//   - si el email ya esta registrado en el mismo proyecto, devuelve
//     app.ErrEmailDuplicado (RN-003-7). El mismo email en otro proyecto es valido;
//   - es de todo o nada: si falla no guarda nada (RN-003-9).
//
// El email llega en minuscula, porque lo normaliza el dominio; igual se compara
// pasando los dos a minuscula, asi la unicidad no depende de como se escribio.
// Con mu tomado, dos altas simultaneas del mismo email no pasan las dos.
func (b *Base) RegistrarIntegrante(_ context.Context, i integrante.Integrante) (integrante.Integrante, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.proyectos[i.ProyectoID] {
		return integrante.Integrante{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, i.ProyectoID)
	}

	email := strings.ToLower(i.Email)
	for _, otro := range b.integrantes[i.ProyectoID] {
		if strings.ToLower(otro.Email) == email {
			return integrante.Integrante{}, fmt.Errorf("%w (email %s)", app.ErrEmailDuplicado, i.Email)
		}
	}

	i.ID = b.nuevoID()
	b.integrantes[i.ProyectoID] = append(b.integrantes[i.ProyectoID], i)
	return i, nil
}

// Integrantes devuelve una copia de los integrantes registrados en el proyecto,
// en orden de alta. Un proyecto sin integrantes, o que no existe, devuelve una
// lista vacia. Lo usan los escenarios para verificar lo que quedo guardado:
// todavia no hay un caso de uso de lectura.
func (b *Base) Integrantes(proyectoID string) []integrante.Integrante {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.integrantes[proyectoID])
}
