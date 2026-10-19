package backlog

import (
	"slices"
	"time"
)

// Completar devuelve una copia del item completado en la fecha recibida (RN-010-5),
// desde pendiente o en progreso. La fecha entra por parametro porque el dominio no lee
// el reloj (RN-010-8).
//
// Marcar un item que ya esta completado no es un error y no cambia nada: conserva la
// fecha original (RN-010-6), asi un doble clic o un reintento no mueven el burndown.
//
// Que el item este en un sprint abierto lo valida el dominio de sprint
// (VerificarMarcado), no este metodo.
func (i ItemBacklog) Completar(en time.Time) ItemBacklog {
	i.Criterios = slices.Clone(i.Criterios)
	if i.Estado == EstadoCompletado {
		return i
	}
	i.Estado = EstadoCompletado
	i.CompletadoEn = en
	return i
}

// Reabrir devuelve una copia del item desmarcado: pendiente y sin fecha de
// completado, aunque antes de completarse estuviera en progreso, porque no se guarda
// el estado anterior (RN-010-7). Desmarcar un item que no esta completado no es un
// error y no cambia nada.
func (i ItemBacklog) Reabrir() ItemBacklog {
	i.Criterios = slices.Clone(i.Criterios)
	if i.Estado != EstadoCompletado {
		return i
	}
	i.Estado = EstadoPendiente
	i.CompletadoEn = time.Time{}
	return i
}
