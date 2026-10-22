// Package aquas embebe el gobierno central para que `aquas install` lo lleve a cada hijo.
// La versión del binario (go install …@vX.Y.Z) es la versión del gobierno que instala.
package aquas

import "embed"

// Central: canónicos y config de gobierno, tal como están en este commit.
//
//go:embed gobierno/canonico gobierno/config
var Central embed.FS
