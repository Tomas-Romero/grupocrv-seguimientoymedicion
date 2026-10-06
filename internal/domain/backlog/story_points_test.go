package backlog_test

import (
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// US-005 / RN-005-7, CL-005-10: el valor cero de StoryPoints es "sin estimar", no
// 0 puntos, y es igual a SinEstimar().
func TestStoryPoints_ValorCeroEsSinEstimar(t *testing.T) {
	var puntos backlog.StoryPoints

	if valor, estimado := puntos.Valor(); estimado || valor != 0 {
		t.Errorf("Valor() = (%d, %t), se esperaba (0, false)", valor, estimado)
	}
	if puntos != backlog.SinEstimar() {
		t.Errorf("el valor cero %+v no es igual a SinEstimar() %+v", puntos, backlog.SinEstimar())
	}
}
