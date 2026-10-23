// Paquete domain: reglas de negocio puras, sin base de datos ni HTTP.
// Al no tener dependencias se pueden probar directamente.
package domain

// Estados posibles de un pedido.
const (
	EstadoPendiente  = "pendiente"
	EstadoPreparando = "preparando"
	EstadoEnviado    = "enviado"
	EstadoEntregado  = "entregado"
	EstadoCancelado  = "cancelado"
)

// transiciones define a qué estados se puede pasar desde cada estado.
// Un pedido entregado o cancelado ya no se mueve: es el final del flujo.
var transiciones = map[string][]string{
	EstadoPendiente:  {EstadoPreparando, EstadoCancelado},
	EstadoPreparando: {EstadoEnviado, EstadoCancelado},
	EstadoEnviado:    {EstadoEntregado, EstadoCancelado},
	EstadoEntregado:  {},
	EstadoCancelado:  {},
}

// EstadoValido indica si el texto corresponde a un estado conocido.
func EstadoValido(estado string) bool {
	_, existe := transiciones[estado]
	return existe
}

// TransicionPermitida dice si se puede pasar de un estado a otro.
func TransicionPermitida(desde, hasta string) bool {
	for _, posible := range transiciones[desde] {
		if posible == hasta {
			return true
		}
	}
	return false
}

// SiguientesEstados devuelve los estados a los que puede avanzar un pedido.
func SiguientesEstados(desde string) []string {
	siguientes := transiciones[desde]
	if siguientes == nil {
		return []string{}
	}
	return siguientes
}
