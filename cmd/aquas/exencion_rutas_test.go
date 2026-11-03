package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestExencionRutasDelSistema (E000/F007/S001): la exención de `aquas memory set` compara rutas como el sistema.
// Ignora mayúsculas solo en Windows y acepta el binario de la raíz en forma Git Bash (/d/…).
// Rojo cuando: /d/ entre comillas no exenta o en Linux otra mayúscula exenta.
// Verde cuando: ambas formas se comportan como el sistema.
func TestExencionRutasDelSistema(t *testing.T) {
	casos := []struct {
		caso, goos, raiz, comando string
		exenta                    bool
	}{
		{"windows: binario en forma Git Bash", "windows", `D:\Repo`, `"/d/Repo/.tmp/aquas.exe" memory set tarea=x`, true},
		{"windows: Git Bash con otras mayúsculas", "windows", `D:\Repo`, `"/d/repo/.tmp/aquas.exe" memory set tarea=x`, true},
		{"windows: forma D:/ con otras mayúsculas", "windows", `D:\Repo`, `"d:/repo/.tmp/aquas.exe" memory set tarea=x`, true},
		{"windows: binario de otra carpeta", "windows", `D:\Repo`, `"/d/Otro/.tmp/aquas.exe" memory set tarea=x`, false},
		{"windows: cd a la raíz con otras mayúsculas", "windows", `D:\Repo`, `cd /d/repo && aquas memory set tarea=x`, true},
		{"windows: cd a otra carpeta", "windows", `D:\Repo`, `cd /d/otro && aquas memory set tarea=x`, false},
		{"linux: cd con otras mayúsculas no es la raíz", "linux", "/home/u/Repo", `cd /home/u/repo && aquas memory set tarea=x`, false},
		{"linux: cd a la raíz exacta", "linux", "/home/u/Repo", `cd /home/u/Repo && aquas memory set tarea=x`, true},
		{"linux: binario con otras mayúsculas no es el de la raíz", "linux", "/home/u/Repo", `"/home/u/REPO/.tmp/aquas" memory set tarea=x`, false},
		{"linux: binario de la raíz", "linux", "/home/u/Repo", `"/home/u/Repo/.tmp/aquas" memory set tarea=x`, true},
		{"linux: /d/ no es una unidad", "linux", "/home/u/Repo", `"/d/Repo/.tmp/aquas" memory set tarea=x`, false},
		{"windows: cd /d con la raíz en la unidad", "windows", `D:\`, `cd /d && aquas memory set tarea=x`, true},
		{"windows: cd /d/ con la raíz en la unidad", "windows", `D:\`, `cd /d/ && aquas memory set tarea=x`, true},
		{"windows: D: a secas no es la raíz de la unidad", "windows", `D:\`, `cd D: && aquas memory set tarea=x`, false},
		{"windows: ruta UNC de la raíz", "windows", `\\srv\share\repo`, `cd //srv/share/repo && aquas memory set tarea=x`, true},
		{"windows: barra simple no es la UNC", "windows", `\\srv\share\repo`, `cd /srv/share/repo && aquas memory set tarea=x`, false},
		{"orden que no es memoria no queda exenta", "windows", `D:\Repo`, `"/d/Repo/.tmp/aquas.exe" approval`, false},
	}
	for _, c := range casos {
		if got := exentaDelFrenoEn(c.goos, c.raiz, "Bash", c.comando); got != c.exenta {
			t.Errorf("%s: exenta=%v, quiere %v", c.caso, got, c.exenta)
		}
	}
	// Choque con la guardia: el binario en forma Git Bash no salta el freno de las órdenes de hook.
	raiz := repoSano(t)
	gitBash := "/" + strings.ToLower(raiz[:1]) + filepath.ToSlash(raiz[2:])
	b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": `"` + gitBash + `/.tmp/aquas.exe" approval`}})
	if s := aquas(t, raiz, string(b), "guard"); s.codigo != 2 {
		t.Errorf("la guardia frena approval aunque el binario venga en forma Git Bash; dio %+v", s)
	}
}
