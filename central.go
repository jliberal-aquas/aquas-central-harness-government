// Package aquas embebe el gobierno central para que `aquas install` lo lleve a cada hijo.
// La versión del binario (go install …@vX.Y.Z) es la versión del gobierno que instala.
package aquas

import "embed"

// Central: canónicos y config versionada de gobierno, tal como están en este commit.
// E003/F001/S001: la config va por nombre; un *.local.json de la máquina que compila no viaja.
// Archivo nuevo en gobierno/config: agrégalo aquí; TestBinarioSinEntornoLocal lo exige.
//
//go:embed gobierno/canonico
//go:embed gobierno/config/agentes.json gobierno/config/claude_settings.json gobierno/config/dependencias.json
//go:embed gobierno/config/documentos.json gobierno/config/golangci.yml gobierno/config/guardia.json
//go:embed gobierno/config/sellos.json
var Central embed.FS
