# Coding Challenges — Algoritmos Idiomaticos en Go

Problemas clasicos de algoritmos y estructuras de datos resueltos de forma **idiomatica en Go**, no "Java traducido a Go".

## Enfoque

En entrevistas con Go, los entrevistadores buscan:

1. **Uso correcto de slices, maps, y la standard library**: no reinventar lo que ya existe.
2. **Codigo limpio y legible**: Go prioriza la claridad sobre la cleverness.
3. **Manejo de edge cases**: nil inputs, slices vacios, strings vacios.
4. **Complejidad algoritmica**: saber explicar Big-O de tiempo y espacio.
5. **Idiomatic Go**: nombrado corto, error handling explicito, zero values utiles.

## Estructura

- **`challenges.go`**: 10 funciones con `panic("TODO")`. Tu tarea es implementarlas.
- **`challenges_test.go`**: tests completos con table-driven tests y edge cases. Ejecuta con `go test`.
- **`solutions.go.txt`**: soluciones completas con explicaciones. Renombralo a `solutions.go` y borra `challenges.go` para verificar que los tests pasan.

## Challenges

| # | Funcion | Descripcion | Dificultad |
|---|---|---|---|
| 1 | `TwoSum` | Encontrar dos numeros que sumen un target | Facil |
| 2 | `ValidParentheses` | Verificar brackets balanceados | Facil |
| 3 | `MergeSortedArrays` | Fusionar dos arrays ordenados | Facil |
| 4 | `ReverseLinkedList` | Invertir una linked list | Facil |
| 5 | `LRUCache` | Implementar cache LRU con Get/Put | Media |
| 6 | `MaxSubarraySum` | Subarray con suma maxima (Kadane) | Media |
| 7 | `BinarySearch` | Busqueda binaria clasica | Facil |
| 8 | `LevelOrderTraversal` | BFS en arbol binario | Media |
| 9 | `IsAnagram` | Verificar si dos strings son anagramas | Facil |
| 10 | `TopKFrequent` | K elementos mas frecuentes | Media |

## Como Usar

```bash
# Ejecutar tests (fallaran hasta que implementes las soluciones)
cd 02-interview-prep/coding-challenges
go test -v

# Ver que tests fallan
go test -v -run TestTwoSum

# Usar las soluciones para verificar
cp solutions.go.txt solutions.go
rm challenges.go
go test -v
```

## Consejos para Entrevistas

- **Empieza por los edge cases**: que pasa con input vacio? nil? un solo elemento?
- **Explica tu enfoque antes de codear**: "Voy a usar un hash map para O(1) lookup..."
- **Menciona la complejidad**: "Esto es O(n) en tiempo y O(n) en espacio."
- **Usa el zero value**: en Go, el zero value de un `map[K]V` al acceder una key inexistente es el zero value de V. Aprovechalo.
- **Nombra bien**: `i, j` para indices, `n` para tamano, `ok` para booleanos de existencia.
