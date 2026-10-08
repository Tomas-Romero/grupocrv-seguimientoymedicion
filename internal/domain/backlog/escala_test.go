package backlog_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// US-013 / RN-013-1: la escala es 1, 2, 3, 5, 8 y 13.
func TestEscalaFibonacci(t *testing.T) {
	esperada := []int{1, 2, 3, 5, 8, 13}
	if escala := backlog.EscalaFibonacci(); !slices.Equal(escala, esperada) {
		t.Errorf("EscalaFibonacci() = %v, se esperaba %v", escala, esperada)
	}
}

// La escala se devuelve como copia: quien la recibe no puede cambiar la regla.
func TestEscalaFibonacci_DevuelveUnaCopia(t *testing.T) {
	escala := backlog.EscalaFibonacci()
	escala[0] = 99

	if otra := backlog.EscalaFibonacci(); otra[0] != 1 {
		t.Errorf("cambiar la copia modifico la escala: EscalaFibonacci()[0] = %d, se esperaba 1", otra[0])
	}
}

// US-013 / CA-013-1, CL-013-1: cada valor de la escala, incluidos los extremos 1
// y 13, es una estimacion valida y queda marcada como estimada.
func TestNuevosStoryPoints_ValoresDeLaEscala(t *testing.T) {
	for _, puntos := range []int{1, 2, 3, 5, 8, 13} {
		t.Run(strconv.Itoa(puntos), func(t *testing.T) {
			sp, err := backlog.NuevosStoryPoints(puntos)
			if err != nil {
				t.Fatalf("NuevosStoryPoints(%d): %v", puntos, err)
			}
			if valor, estimado := sp.Valor(); valor != puntos || !estimado {
				t.Errorf("Valor() = (%d, %t), se esperaba (%d, true)", valor, estimado, puntos)
			}
		})
	}
}

// US-013 / RN-013-1, RN-013-2, CL-013-2 a CL-013-5, CA-013-4: cualquier valor
// fuera de la escala se rechaza, incluidos el 0, los negativos, los que no son
// Fibonacci y los Fibonacci que la escala no incluye (21 y 34). El error indica
// el valor recibido.
func TestNuevosStoryPoints_FueraDeLaEscala(t *testing.T) {
	for _, puntos := range []int{0, -3, 4, 6, 7, 9, 10, 12, 14, 21, 34, 100} {
		t.Run(strconv.Itoa(puntos), func(t *testing.T) {
			sp, err := backlog.NuevosStoryPoints(puntos)
			if !errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
				t.Fatalf("error = %v, se esperaba ErrEstimacionFueraDeEscala", err)
			}
			if sp != backlog.SinEstimar() {
				t.Errorf("devolvio %+v, se esperaba SinEstimar()", sp)
			}
			if esperado := strconv.Itoa(puntos); !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error %q no indica el valor recibido %s", err, esperado)
			}
		})
	}
}
