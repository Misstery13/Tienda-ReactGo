// Reparto de comisiones del plan multinivel.
// Función pura: sin base de datos ni HTTP, se puede probar directamente.
package domain

import "math"

// TasasPorNivel define cuánto gana cada ancestro del comprador:
// el primero arriba 10 %, el segundo 5 %, el tercero 2 %.
var TasasPorNivel = []float64{0.10, 0.05, 0.02}

// Comision es una línea del reparto, todavía sin dueño asignado.
type Comision struct {
	Nivel      int     `json:"nivel"`
	Porcentaje float64 `json:"porcentaje"`
	Monto      float64 `json:"monto"`
}

// RepartirComisiones calcula cuánto corresponde a cada nivel sobre un total.
// Devuelve una línea por tasa; los montos van redondeados a centavos.
func RepartirComisiones(total float64, tasas []float64) []Comision {
	comisiones := make([]Comision, 0, len(tasas))

	if total <= 0 {
		return comisiones
	}

	for i, tasa := range tasas {
		if tasa <= 0 {
			continue
		}
		comisiones = append(comisiones, Comision{
			Nivel:      i + 1,
			Porcentaje: tasa,
			Monto:      math.Round(total*tasa*100) / 100,
		})
	}

	return comisiones
}
