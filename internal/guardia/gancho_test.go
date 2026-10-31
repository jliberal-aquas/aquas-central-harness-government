package guardia

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestGanchoAManoBloqueado (E000/F008/S001, D000077): la guardia bloquea por Bash las órdenes
// reservadas a los hooks y deja pasar las libres.
// Rojo cuando: aquas approval por Bash pasa.
// Verde cuando: cada orden de hook da exit 2 y memory set pasa.
func TestGanchoAManoBloqueado(t *testing.T) {
	raiz := raizPrueba(t)
	bloquea := []string{
		`echo '{}' | ./.tmp/aquas.exe approval`,
		`./.tmp/aquas.exe agent-close < x.json`,
		`./.tmp/aquas.exe budget prompt`,
		`aquas budget tool`,
		`aquas turn open`,
		`aquas proposal`,
		`aquas guard`,
		`aquas hook`,
		`aquas memory start`,
		`aquas memory close`,
		// nombres viejos
		`aquas aprobacion`,
		`aquas cierre-agente`,
		`aquas presupuesto prompt`,
		`aquas turno cerrar`,
		`aquas propuesta`,
		`aquas guardia`,
		`aquas memoria inicio`,
		`aquas memoria cierre`,
		// formas de llamar
		`"D:/Agents Factory/x/.tmp/aquas.exe" approval`,
		`./.tmp/aq.exe approval`,
		`sh .claude/hooks/aquas.sh approval`,
		`sh "x/aquas.sh" agent-close`,
		`go run ./cmd/aquas approval`,
		`go run ./cmd/aquas/ approval`,
		`X=1 aquas approval`,
		`cd docs && aquas approval`,
		`echo a; aquas approval`,
		`sh -c "aquas approval"`,
		`bash -c 'echo x | ./.tmp/aquas.exe budget prompt'`,
		`C:\x\aquas.exe approval`,
		// A1: nombres cruzados, redirección delante, prefijos, $(…), -lc, go run, variable
		`aquas memory inicio`,
		`aquas memoria start`,
		`aquas memoria close`,
		`< falso.json aquas approval`,
		`<falso.json ./.tmp/aquas.exe agent-close`,
		`2>/dev/null aquas approval`,
		`aquas 2>/dev/null approval`,
		`echo $(aquas approval)`,
		`x=$(aquas approval)`,
		`env aquas approval`,
		`exec aquas approval`,
		`nohup aquas approval`,
		`time aquas approval`,
		`command aquas approval`,
		`! aquas approval`,
		`{ aquas approval; }`,
		`if true; then aquas approval; fi`,
		`bash -lc "aquas approval"`,
		`sh -xc 'aquas approval'`,
		`go run -tags x ./cmd/aquas approval`,
		`cd cmd/aquas && go run . approval`,
		`go run ./cmd/aquas/main.go approval`,
		`powershell.exe -Command "aquas approval"`,
		`cmd //c aquas approval`,
		`a=aquas; $a approval`,
		`aquas APPROVAL`,
		// A2: redirección pegada detrás, comillas y escapes dentro de la orden o del nombre
		`./.tmp/aquas.exe agent-close<falso.json`,
		`aquas approval<falso.json`,
		`aquas approval>.tmp/x`,
		`aquas memory start<x`,
		`aquas "approval"`,
		`aquas 'approval'`,
		`aquas app""roval`,
		`aquas ap\proval`,
		`aquas memory 'start'`,
		`aq""uas approval`,
		`a'qu'as approval`,
		"aquas \\\napproval",
		// A3: here-string no es heredoc; las líneas siguientes se revisan
		`aquas <<<x approval`,
		"cat <<<x\naquas approval",
		// T1: «<<» entre comillas no es heredoc
		"git commit -m \"doc: cat <<EOF\"\naquas approval",
		// A5: comilla abierta entre líneas, aritmética sin $, delimitador con guion o comillas
		"echo \"\nx <<FIN\n\"\naquas approval\nFIN",
		"echo '\n<<FIN\n'\naquas approval\nFIN",
		"((y=1<<z))\naquas approval\nz",
		"for ((i=0; i<<n; i++)); do :; done\naquas approval\nn",
		"cat <<END-OF\nEND-OF\naquas approval\nEND",
		"cat <<E\"O\"F\nEOF\naquas approval\nE",
		// A6: $'…', # tras ), comillas dentro de $((…)), delimitador con blanco
		"echo $'\\' <<FIN\n'\naquas approval\nFIN",
		"echo $(echo x)#'\n<<FIN\n'\naquas approval\nFIN",
		"echo $(( $(echo 1; : \"))\") <<FIN ))\naquas approval\nFIN",
		"cat <<'FIN ' > o4.txt\nx\nFIN \naquas approval",
		// A7: heredoc que lee un shell o que sigue por tubería
		"sh <<FIN\naquas approval\nFIN",
		"bash -s <<\"FIN\"\naquas approval\nFIN",
		"source /dev/stdin <<FIN\naquas approval\nFIN",
		"cat <<FIN | sh\naquas approval\nFIN",
		// A8: shell dentro de comillas, cat redefinido antes
		"sh -c 'sh; cat x' <<FIN\naquas approval\nFIN",
		"bash -c 'read l; eval $l; cat x' <<FIN\naquas approval\nFIN",
		"function cat { bash; }\ncat <<FIN\naquas approval\nFIN",
		"shopt -s expand_aliases\nalias cat=sh\ncat <<FIN\naquas approval\nFIN",
		"alias cat=sh; cat <<FIN\naquas approval\nFIN",
		// revisa de más (precio aceptado): heredoc que no entiende por completo
		"cat <<E\"O\"F >x.md\naquas approval citado\nEOF",
		"cat <<A >a.md; cat <<B >b.md\naquas approval\nA\nB",
		// precio aceptado: texto entre comillas que cita la orden
		`echo "aquas approval"`,
	}
	for _, c := range bloquea {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "solo la lanza Claude Code") {
			t.Errorf("%q debe bloquearse como orden de hook; dio %q", c, m)
		}
	}
	pasa := []string{
		`./.tmp/aquas.exe memory set tarea="x"`,
		`aquas memory show`,
		`aquas memory archive`,
		`aquas backlog show E1`,
		`aquas defect new firma=a`,
		`./.tmp/aquas.exe verify`,
		`go run ./cmd/aquas compile`,
		`grep approval cmd/aquas/main.go`,
		`go test ./internal/guardia`,
		`go run ./cmd/aquas defect new que="la guardia del turn" causa=x`,
		`go run ./cmd/aquas verify && git commit -m "fix the guard"`,
		`git log --grep approval`,
		`echo $HOME`,
		"git commit -q -F - <<'EOF'\naquas approval citado en el mensaje\nEOF",
		"cd \"docs\" && cat >n.md <<-FIN\n\taquas approval citado\n\tFIN",
	}
	for _, c := range pasa {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q debe pasar; dio %q", c, m)
		}
	}
	// A3: la regla de oro tampoco pierde las líneas que siguen a un here-string.
	if m := RevisarBash(raiz, raiz, "cat <<<x\nrm -rf /c/Windows/x"); !strings.Contains(m, "fuera del repositorio") {
		t.Errorf("here-string seguido de escritura fuera: debe bloquearse; dio %q", m)
	}
	// T1: aritmética, comillas y comentario no abren heredoc; las líneas siguientes se revisan.
	for _, c := range []string{
		"echo $((a<<b))\nrm -rf /c/Windows/x",
		"echo \"usa <<FIN para heredocs\"\nrm -rf /c/Windows/x",
		"echo x # <<FIN\nrm -rf /c/Windows/x",
		"echo \"\nx <<FIN\n\"\nrm -rf /c/Windows/x\nFIN",
		"((y=1<<z))\nrm -rf /c/Windows/x\nz",
		"cat <<END-OF\nEND-OF\nrm -rf /c/Windows/x\nEND",
		"sh <<FIN\nrm -rf /c/Windows/x\nFIN",
		"sh -c 'sh; cat x' <<FIN\nrm -rf /c/Windows/x\nFIN",
		"function cat { bash; }\ncat <<FIN\nrm -rf /c/Windows/x\nFIN",
		"echo $'\\' <<FIN\n'\nrm -rf /c/Windows/x\nFIN",
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "fuera del repositorio") {
			t.Errorf("%q debe bloquearse como fuera del repositorio; dio %q", c, m)
		}
	}
	// Quién habla: un subagente, con agent_type, también se bloquea.
	e, _ := json.Marshal(map[string]any{"tool_name": "Bash", "agent_type": "worker-sonnet", "tool_input": map[string]string{"command": "aquas approval"}})
	if m := Revisar(raiz, e); !strings.Contains(m, "solo la lanza Claude Code") {
		t.Errorf("subagente: debe bloquearse; dio %q", m)
	}
}
