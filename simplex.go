package linearprog

import (
	"math"
	"runtime"
)

func Round(number float64, precision int) float64 {
	var out float64
	v1 := math.Pow(10, float64(precision))
	v2 := int(number*v1 + math.Copysign(0.5, number*v1))
	out = float64(v2) / v1
	return out
}

//Aqui se puede mejorar si se consigue una recurrencia de los 0, cada cuanto hay cero, entonces no sumarlo.
func DifferenceRows(Pivot map[int]map[int]float64, rowsids []int, colpivot, rowspivot, columns int, c chan map[int]map[int]float64) {
	Pivot2 := make(map[int]map[int]float64)
	numberrows := len(rowsids)
	for i := 0; i < numberrows; i++ {
		lowervalue := Pivot[rowsids[i]][colpivot]
		if rowsids[i] != rowspivot && lowervalue != 0 {
			for j := 0; j < columns; j++ {
				if Pivot2[rowsids[i]] == nil {
					Pivot2[rowsids[i]] = map[int]float64{}
				}
				Pivot2[rowsids[i]][j] = (-lowervalue)*(Pivot[rowspivot][j]) + Pivot[rowsids[i]][j]
			}
		}
	}
	c <- Pivot2
}

//func InitMap(rows, columns int) map[int]map[int]float64 {
func InitMap(rows int) map[int]map[int]float64 {
	Pivot := make(map[int]map[int]float64)

	for i := 0; i < rows; i++ {
		Pivot[i] = map[int]float64{}
	}

	// for i := 0; i < rows; i++ {
	// 	Pivot[i] = map[int]float64{}
	// 	for j := 0; j < columns; j++ {
	// 		Pivot[i][j] = 0
	// 	}
	// }
	return Pivot
}

func PartitionForGorutine(Pivot map[int]map[int]float64) (float64, int) {
	n := runtime.NumCPU()
	var N float64
	if len(Pivot) >= n {
		N = math.Floor(float64(len(Pivot)) / float64(n))
	} else {
		N = float64(len(Pivot)) / float64(n)
	}
	return N, n
}

func Simplex(A map[int]map[int]float64, b []float64, a []float64, constdir []string) (map[int]float64, float64) {
	var feasible bool
	var rows, columns int
	solutions := make(map[int]float64)
	rows = len(b) + 1
	columns = len(a) + len(b) + 2

	//Pivot := InitMap(rows, columns)
	Pivot := InitMap(rows)
	//Estructura del simplex
	Pivot[0][0] = 1.00
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			if i == 0 && j >= 1 && j <= len(a) {
				Pivot[i][j] = -a[j-1]
			} else if i > 0 && j >= 1 && j <= len(a) {
				Pivot[i][j] = A[i][j]
			} else if i > 0 && j > len(a) && j < columns-1 && constdir[i-1] == "<=" {
				Pivot[i][i+len(a)] = 1.00
			} else if i > 0 && j > len(a) && j < columns-1 && constdir[i-1] == ">=" {
				Pivot[i][i+len(a)] = -1.00
			} else if i > 0 && j == columns-1 {
				Pivot[i][j] = b[i-1]
			}
		}
	}

	N, n := PartitionForGorutine(Pivot)

	r := 0.00
	var rowsids [][]int
	var ids []int
	s1, s2 := 0.00, 1.00

	for i := range Pivot {
		if s1 < (s2*N + r) {
			ids = append(ids, int(i))
			if s1 == ((float64(n) * N) + r - 1) {
				rowsids = append(rowsids, ids)
			}
		} else {
			s2++
			if s2 == float64(n) {
				r = math.Mod(float64(len(Pivot)), float64(n))
			}
			rowsids = append(rowsids, ids)
			ids = []int{i}
			if s1 == ((float64(n) * N) + r - 1) {
				rowsids = append(rowsids, ids)
			}
		}
		s1++
	}

	if len(rowsids) < n {
		n = len(rowsids)
	}

	//se consigue la fila y columna pivote, para tener el elementpivote
	for {
		colpivot := 1
		min := Pivot[0][1]
		for i := 1; i < columns; i++ {
			if feasible {
				if Pivot[0][i] < 0 && min > Pivot[0][i] {
					colpivot = i
					min = Pivot[0][i]
				}
			} else {
				if Pivot[0][i] < 0 && min > Pivot[0][i] && i <= len(a) {
					colpivot = i
					min = Pivot[0][i]
				}
			}
		}

		min = math.Inf(1)
		rowspivot := 1
		for k := 1; k < rows; k++ {
			v := float64(Pivot[k][columns-1]) / float64(Pivot[k][colpivot])
			if v > 0 {
				if min > v {
					min = v
					rowspivot = k
				}
			}
		}
		/////////////////////////////
		elementpivot := Pivot[rowspivot][colpivot]
		//fmt.Println("elmento", elementpivot)
		for j := 0; j < columns; j++ {
			Pivot[rowspivot][j] = float64(Pivot[rowspivot][j]) / float64(elementpivot)
			//fmt.Println("cociente", Pivot[rowspivot][j])
		}
		///////////////////////////mejorar ver como pasar el tipo de datos anteriores a funciones
		c := make(chan map[int]map[int]float64)
		for i := 0; i < n; i++ {
			go DifferenceRows(Pivot, rowsids[i], colpivot, rowspivot, columns, c)
		}

		u := make([]map[int]map[int]float64, n)
		for i := 0; i < n; i++ {
			u[i] = <-c
		}

		for i := 0; i < n; i++ {
			for k := range u[i] {
				Pivot[k] = u[i][k]
			}
		}

		number, numberfeasible := 0, 0
		// for i := 0; i < columns; i++ {
		// 	if Pivot[0][i] >= 0 {
		// 		number = number + 1
		// 	}
		// 	if Pivot[0][i] >= 0 && i < len(a) {
		// 		numberfeasible = numberfeasible + 1
		// 	}
		// }

		for i := 0; i < columns; i++ {
			if Pivot[0][i] >= 0 {
				number = number + 1
			} else {
				numberfeasible = numberfeasible + 1
			}
		}

		if numberfeasible == len(a) {
			feasible = true
		}
		if number == columns {
			break
		}
	}

	opt := Pivot[0][columns-1]
	//checar como evitar este ultimo paso con la solucion, ver por donde viene o previamente sacarla

	//Se puede guardar la columna que tiene solo 1 y ceros, de ahi sacamos la fila y la solucion final de esa variable sera: Solucion[columna guardada digamos la 2]= resultado de la ultima columna de la fila que tiene el 1.
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if Pivot[0][i] > 0 {
				solutions[i-1] = 0
			} else if Pivot[0][i] == 0 && Pivot[j][i] == 1 {
				solutions[i-1] = Round(Pivot[j][columns-1], 2)
			}
		}
	}

	return solutions, opt

}
