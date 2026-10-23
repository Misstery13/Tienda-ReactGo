package domain

import "testing"

func TestRepartirComisiones(t *testing.T) {
	casos := []struct {
		nombre   string
		total    float64
		tasas    []float64
		esperado []Comision
	}{
		{
			nombre: "reparto normal de tres niveles",
			total:  100,
			tasas:  TasasPorNivel,
			esperado: []Comision{
				{Nivel: 1, Porcentaje: 0.10, Monto: 10},
				{Nivel: 2, Porcentaje: 0.05, Monto: 5},
				{Nivel: 3, Porcentaje: 0.02, Monto: 2},
			},
		},
		{
			nombre:   "total cero no genera comisiones",
			total:    0,
			tasas:    TasasPorNivel,
			esperado: []Comision{},
		},
		{
			nombre:   "total negativo no genera comisiones",
			total:    -50,
			tasas:    TasasPorNivel,
			esperado: []Comision{},
		},
		{
			nombre: "los montos se redondean a centavos",
			total:  118.33,
			tasas:  []float64{0.10},
			esperado: []Comision{
				{Nivel: 1, Porcentaje: 0.10, Monto: 11.83},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			obtenido := RepartirComisiones(caso.total, caso.tasas)

			if len(obtenido) != len(caso.esperado) {
				t.Fatalf("se esperaban %d comisiones, llegaron %d", len(caso.esperado), len(obtenido))
			}

			for i, esperada := range caso.esperado {
				if obtenido[i] != esperada {
					t.Errorf("nivel %d: se esperaba %+v, llegó %+v", i+1, esperada, obtenido[i])
				}
			}
		})
	}
}

func TestTransicionPermitida(t *testing.T) {
	casos := []struct {
		desde, hasta string
		permitida    bool
	}{
		{EstadoPendiente, EstadoPreparando, true},
		{EstadoPendiente, EstadoCancelado, true},
		{EstadoPreparando, EstadoEnviado, true},
		{EstadoEnviado, EstadoEntregado, true},
		{EstadoEntregado, EstadoPendiente, false}, // no se retrocede
		{EstadoEntregado, EstadoCancelado, false}, // entregado es final
		{EstadoCancelado, EstadoCancelado, false}, // no se cancela dos veces
		{EstadoPendiente, EstadoEntregado, false}, // no se saltan pasos
	}

	for _, caso := range casos {
		if obtenido := TransicionPermitida(caso.desde, caso.hasta); obtenido != caso.permitida {
			t.Errorf("de %q a %q: se esperaba %v, llegó %v",
				caso.desde, caso.hasta, caso.permitida, obtenido)
		}
	}
}
