---
PLAN: "feat: kind Vector(dim) para columnas de embeddings"
TAG: v0.2.0
EXECUTOR: unassigned
REVIEWER: none
---

> Parte del esfuerzo de búsqueda semántica nativa en el navegador. Índice maestro:
> https://github.com/webtyp/agent/blob/main/docs/PLAN.md — la decisión **D1** de ahí es la
> justificación de todo lo de abajo y no se vuelve a argumentar acá.
>
> **Nota de idioma:** la prosa va en español; los bloques de código mantienen sus
> comentarios en inglés, como el resto del código fuente de este repositorio.

# Plan — un kind `Vector(dim)`

## Por qué

La búsqueda semántica necesita una columna que guarde un embedding: un `[]float32` de largo
fijo serializado little-endian. `model` ya puede transportar los bytes — `FieldBlob` existe
(`field.go:13`) y está cableado en `IsZeroPtr` (`field.go:310`) y `ValuesFrom`
(`field.go:383`) — pero una columna `Blob()` a secas **no lleva dimensión**, así que nada
aguas abajo puede rechazar un vector de 384 dims escrito en una columna de 768 dims. Ese
desajuste es silencioso, y corrompe todos los resultados de búsqueda posteriores en vez de
fallar ruidosamente.

Este plan agrega la dimensión al esquema, donde el resto del ecosistema puede leerla.

## Lo que NO cambia

**Ningún `FieldType` nuevo.** `FieldBlob` sigue siendo el tipo de almacenamiento. Agregar
una constante `FieldFloat32Slice` obligaría a editar cada `switch` exhaustivo sobre
`FieldType` en este repositorio, en `storage/mem`, `sqlt`, `postgres` e `indexdb` — un
cambio incompatible en seis repositorios que no compra nada que `FieldBlob` no provea ya.
`Vector(dim)` es un `Kind`, y `Kind` es exactamente la juntura diseñada para "mismo
almacenamiento, distinta semántica" (es lo que hoy distingue a `Text()` de un kind de
email).

La interfaz `Kind` **no** se modifica. Sus tres métodos quedan como están.

`ValuesFrom`, `IsZeroPtr`, `ScanFields` y los codecs no necesitan edición: despachan sobre
`Field.Type.Storage()`, que para un vector devuelve `FieldBlob`, ya contemplado.

## Cambios

### 1. `kind.go` — el constructor `Vector`

Agregar debajo de `Blob()`:

```go
// Dimensional is implemented by kinds whose value has a fixed element count.
// Consumers type-assert for it; a kind that does not implement it is unconstrained.
type Dimensional interface {
	Dim() int
}

type vectorKind struct {
	baseKind
	dim int
}

func (k vectorKind) Dim() int { return k.dim }

// Vector returns a kind for a fixed-length float32 embedding, stored as a
// little-endian blob of dim*4 bytes.
//
// Kind.Validate operates on a string and cannot see the bytes, so it always
// passes here; dimension enforcement is ValidateVector, called by the storage
// layer on the []byte itself.
func Vector(dim int) Kind {
	return vectorKind{
		baseKind: baseKind{
			storage: FieldBlob,
			name:    "vector",
			valid:   func(string) error { return nil },
		},
		dim: dim,
	}
}
```

`vectorKind` embebe a `baseKind`, así que satisface `Kind` sin cuerpos de método extra.

### 2. `field.go` — validación a nivel de bytes

`Kind.Validate(value string) error` recibe un string, que no puede expresar una restricción
sobre un blob. En vez de ensanchar esa interfaz (la implementa cada kind y se la llama desde
`form`, `json` y `orm`), agregar una función libre:

```go
// ValidateVector checks that b is a well-formed value for field f: a multiple of
// four bytes, and exactly f.Type.Dim()*4 bytes when the kind declares a dimension.
// A nil or empty b is accepted for a nullable field and rejected when f.NotNull.
func ValidateVector(f Field, b []byte) error {
	if len(b) == 0 {
		if f.NotNull {
			return fmt.Err("field", f.Name, "is required")
		}
		return nil
	}
	if len(b)%4 != 0 {
		return fmt.Err("field", f.Name, "vector length", len(b), "is not a multiple of 4")
	}
	if d, ok := f.Type.(Dimensional); ok && len(b)/4 != d.Dim() {
		return fmt.Err("field", f.Name, "expects", d.Dim(), "dimensions, got", len(b)/4)
	}
	return nil
}
```

Llamadores: `vectordb` antes de cada escritura, y la suite de conformance de `storage`.

### 3. `field.go` — documentar el mapeo

Extender la tabla de almacenamiento → tipo Go del comentario de doc de `Field` (alrededor
de la línea 72) con una fila para el kind vector:

```
// | FieldBlob (kind "vector") | []byte — dim*4 little-endian float32 |
```

### 4. `docs/` — sin documento nuevo

El contrato de dimensión queda documentado en los comentarios de arriba. La tabla de tipos
del README gana la misma fila.

## Tests

Al estilo de `tests/field_test.go` y `tests/kind_permitted_override_test.go`, solo librería
estándar:

| Test | Verifica |
|---|---|
| `TestVector_StorageIsBlob` | `Vector(384).Storage() == FieldBlob` y `Name() == "vector"` |
| `TestVector_Dim` | el kind satisface `Dimensional` y reporta la dimensión construida |
| `TestVector_ValidateOK` | `ValidateVector` acepta exactamente `dim*4` bytes |
| `TestVector_ValidateWrongDim` | 383 y 385 dims se rechazan ambos, con el nombre del campo en el mensaje |
| `TestVector_ValidateNotMultipleOfFour` | un blob de 1537 bytes se rechaza |
| `TestVector_ValidateEmpty` | vacío pasa cuando es nullable, falla cuando es `NotNull` |
| `TestVector_ZeroPtr` | `IsZeroPtr` sobre un `*[]byte` sigue comportándose para un campo vector |
| `TestBlob_StillNotDimensional` | `Blob()` **no** satisface `Dimensional` — la vía de escape sobrevive |

## Checklist de aceptación

```bash
grep -n "func Vector" kind.go            # → 1 coincidencia
grep -n "func ValidateVector" field.go   # → 1 coincidencia
grep -c "FieldFloat32\|FieldVector" *.go # → 0: no se introdujo ningún FieldType nuevo
go vet ./...
gotest
```

Después liberar, porque `storage` e `indexdb` dependen ambos de este tag:

```bash
gopush 'feat: Vector(dim) kind for embedding columns'
```
