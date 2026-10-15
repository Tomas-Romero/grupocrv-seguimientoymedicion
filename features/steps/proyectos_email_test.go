package steps

import (
	"net/mail"
	"strings"
	"testing"
)

// TestEmailDeNombre verifica las dos promesas de emailDeNombre: el resultado es
// una direccion simple valida para net/mail y es distinto para cada nombre
// distinto. Los nombres cubren espacios, tildes, mayusculas y el nombre vacio.
func TestEmailDeNombre(t *testing.T) {
	nombres := []string{
		"", " ", "   ",
		"Ana", "ana", "ANA",
		"Tomás", "Tomas",
		"Juan Ignacio", "Juan  Ignacio", " Juan Ignacio", "Juan Ignacio ",
		"Ñandú", "Nandu",
		"María José", "O'Brien",
		"a.b", "a_b", "a b", "ab", "x@y",
		"e1", "á", "_e1_",
	}

	vistos := make(map[string]string, len(nombres))
	for _, nombre := range nombres {
		email := emailDeNombre(nombre)

		direccion, err := mail.ParseAddress(email)
		if err != nil {
			t.Errorf("emailDeNombre(%q) = %q: net/mail lo rechaza: %v", nombre, email, err)
			continue
		}
		if direccion.Address != email || direccion.Name != "" {
			t.Errorf("emailDeNombre(%q) = %q: no es una direccion simple (net/mail lee %q)", nombre, email, direccion.Address)
		}
		if email != strings.ToLower(email) {
			t.Errorf("emailDeNombre(%q) = %q: tiene mayusculas y el dominio lo pasaria a minuscula", nombre, email)
		}
		if otro, repetido := vistos[email]; repetido {
			t.Errorf("emailDeNombre(%q) y emailDeNombre(%q) dan el mismo email %q", nombre, otro, email)
		}
		vistos[email] = nombre
	}
}
