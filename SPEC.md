# Hilo — Especificación Técnica

> **Propósito:** TUI (Terminal User Interface) tipo HTTPie/Postman para realizar peticiones HTTP desde la terminal con fluidez, colecciones versionadas, historial y gestión de entornos.

---

## 1. Arquitectura General

### 1.1 Stack Tecnológico
- **Lenguaje:** Go
- **Framework TUI:** Bubble Tea (`charmbracelet/bubbletea`)
- **Estilos:** Lipgloss (`charmbracelet/lipgloss`)
- **HTTP:** `net/http` estándar + `net/http/httputil` para dump de requests/responses
- **Versionado:** Git embebido via `go-git` (`go-git/go-git/v5`) — sin depender de git CLI
- **Almacenamiento local:** Sistema de archivos en `~/.config/hilo/`

### 1.2 Estructura de Archivos

```
~/.config/hilo/
├── config.json                 # Tema, color, modo, último entorno activo
├── collections/                # Colecciones versionadas con git
│   └── mi-api/
│       ├── .git/               # Repo git interno
│       ├── collection.json     # Metadata: nombre, descripción, variables del entorno
│       └── requests/
│           ├── req_001.json
│           ├── req_002.json
│           └── ...
├── history/                    # Historial plano de peticiones
│   ├── 2026-06-01T10-30-00.json
│   └── ...
└── environments/               # Entornos compartidos entre colecciones
    ├── dev.json
    ├── staging.json
    └── prod.json
```

### 1.3 Modelo de Datos

```go
type Request struct {
    ID          string            `json:"id"`
    Collection  string            `json:"collection,omitempty"`
    Name        string            `json:"name"`
    Method      string            `json:"method"`       // GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
    URL         string            `json:"url"`
    Headers     map[string]string `json:"headers,omitempty"`
    QueryParams map[string]string `json:"query_params,omitempty"`
    Body        string            `json:"body,omitempty"`
    BodyType    string            `json:"body_type"`    // none, json, form, text, xml, file
    Auth        *Auth             `json:"auth,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type Auth struct {
    Type     string `json:"type"`     // none, bearer, basic, digest, apikey, oauth2
    Key      string `json:"key,omitempty"`
    Value    string `json:"value,omitempty"`
    Username string `json:"username,omitempty"`
    Password string `json:"password,omitempty"`
}

type Response struct {
    RequestID    string            `json:"request_id"`
    StatusCode   int               `json:"status_code"`
    StatusText   string            `json:"status_text"`
    Headers      map[string]string `json:"headers"`
    Body         string            `json:"body"`
    BodySize     int64             `json:"body_size"`
    Duration     time.Duration     `json:"duration_ms"`
    Timestamp    time.Time         `json:"timestamp"`
    Error        string            `json:"error,omitempty"`
}

type Environment struct {
    Name    string            `json:"name"`
    Values  map[string]string `json:"values"`
}

type Collection struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Environment string `json:"environment,omitempty"`
    Requests    []string `json:"requests"` // ordered list of request IDs
}
```

---

## 2. Interfaz de Usuario (TUI)

### 2.1 Layout General

```
┌──────────────────────────────────────────────────────────────┐
│  [request]  [collection]  [history]  [environments] [config] │  ← Tabs
├──────────────────────────────────────────────────────────────┤
│ ┌─ Method ───┐ ┌──────────── URL ──────────────────────────┐│
│ │ GET  ▼     │ │ https://api.example.com/v1/users          ││
│ └────────────┘ └────────────────────────────────────────────┘│
│ ┌─ Params ──┐ ┌─ Headers ──┐ ┌─ Auth ──┐ ┌─ Body ──┐      │  ← Sub-tabs
│ │           │ │            │ │        │ │        │          │
│ │ key=val   │ │ Accept:json│ │ Bearer │ │ { ... }│          │
│ │ page=1    │ │            │ │        │ │        │          │
│ └───────────┘ └────────────┘ └────────┘ └────────┘          │
│                                                              │
│ [▶ Send]  [Save]  [Copy Curl]  [History]                    │
│                                                              │
│ ┌────────────────── Response ───────────────────────────────┐│
│ │ Status: 200 OK  │  342ms  │  1.2 KB                       ││
│ ├───────────────────────────────────────────────────────────┤│
│ │ {                                                         ││
│ │   "id": 42,                                               ││
│ │   "name": "John Doe",                                     ││
│ │   "email": "john@example.com"                             ││
│ │ }                                                         ││
│ └───────────────────────────────────────────────────────────┘│
│ [Pretty] [Raw] [Headers] [Cookies]                           │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Pantallas / Tabs

| Tab | Propósito |
|-----|-----------|
| **Request** | Panel principal para construir y enviar peticiones HTTP |
| **Collections** | Navegador de colecciones con árbol de requests |
| **History** | Historial cronológico de peticiones enviadas |
| **Environments** | Gestión de variables de entorno (dev/staging/prod) |
| **Config** | Configuración actual de tema, color, modo |

### 2.3 Request Editor — Sub-secciones

Navegación horizontal con tabs dentro del panel Request:

1. **Params** — Query parameters (key=value)
2. **Headers** — Custom headers
3. **Auth** — Autenticación (None, Bearer, Basic, Digest, API Key, OAuth2)
4. **Body** — Cuerpo de la petición (none, JSON, form-data, x-www-form-urlencoded, text, XML, binary)

### 2.4 Response Viewer — Modos de Visualización

- **Pretty** — JSON formateado con sintaxis coloreada
- **Raw** — Body crudo sin formatear
- **Headers** — Headers de respuesta
- **Cookies** — Cookies Set-Cookie

### 2.5 Atajos de Teclado

| Tecla | Acción |
|-------|--------|
| `Ctrl+S` | Enviar petición |
| `Ctrl+E` | Cambiar entre sub-tabs del editor |
| `Ctrl+N` | Nueva petición |
| `Ctrl+D` | Duplicar petición actual |
| `Tab` / `Shift+Tab` | Navegar campos |
| `↑`/`↓` | Navegar listas |
| `Enter` | Seleccionar / abrir |
| `Esc` | Volver / cerrar panel |
| `/` | Buscar en respuesta |
| `Ctrl+C` | Cancelar petición en curso |
| `q` | Salir (confirmar si hay cambios sin guardar) |

---

## 3. Funcionalidades Core

### 3.1 Peticiones HTTP
- Soportar todos los métodos: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- Seguir redirecciones (configurable: follow / no-follow)
- Timeout configurable por petición
- Soporte HTTP/1.1 y HTTP/2
- Certificados TLS: verificación on/off, certificados personalizados
- Proxy HTTP/HTTPS configurable
- Cookie handling (jar opcional)
- Compresión: Accept-Encoding automático

### 3.2 Variables y Entornos
- Variables del estilo `{{BASE_URL}}`, `{{TOKEN}}`, `{{USER_ID}}`
- Sustitución en URL, headers, body, query params, auth
- Entornos separados: dev, staging, production
- Herencia de variables (entorno base + overrides)
- Alternar entre entornos con un atajo

### 3.3 Autenticación

| Tipo | Detalle |
|------|---------|
| None | Sin auth |
| Bearer Token | Header `Authorization: Bearer <token>` |
| Basic Auth | Header `Authorization: Basic <base64>` |
| Digest Auth | Digest Access Authentication |
| API Key | Custom header/query con key=value |
| OAuth2 | Client Credentials flow (grant type, client_id, client_secret, scope, token endpoint) |

### 3.4 Body Editor
- **none** — Sin cuerpo
- **JSON** — Editor con syntax highlighting básico
- **Form-Data** — multipart/form-data (key=value, file upload)
- **x-www-form-urlencoded** — URL encoded key=value
- **Text** — Texto plano
- **XML** — XML plano
- **Binary** — Seleccionar archivo del sistema

### 3.5 Code Snippet Generation
- Generar comando `curl` equivalente
- Opcional: generar snippet en otros lenguajes (Python, Go, JS, etc.)

---

## 4. Colecciones Versionadas (Git)

### 4.1 Estructura Git Interna
Cada colección es un repo git independiente en `~/.config/hilo/collections/<name>/`.

### 4.2 Operaciones Soportadas

| Operación | Descripción |
|-----------|-------------|
| `git init` | Al crear colección |
| `git add .` | Al guardar cambios |
| `git commit -m "msg"` | Commit automático con mensaje descriptivo |
| `git log --oneline` | Ver historial de cambios de la colección |
| `git diff` | Ver diferencias entre versiones de un request |
| `git checkout <hash>` | Restaurar un request a una versión anterior |
| `git branch` | Ramas para experimentar sin afectar la colección principal |

### 4.3 Commits Automáticos
Cada cambio en un request genera un commit con mensaje estructurado:
```
tipo(scope): mensaje corto

tipos: add, update, delete, rename
scope: request, collection, env
```

Ejemplos:
```
add(request): GET /users list endpoint
update(request): add pagination params to /users
delete(request): remove deprecated /v1/old-endpoint
update(collection): rename "My API" to "Production API"
```

### 4.4 Integración Git en TUI
- **Log viewer** — `git log` dentro de la TUI con navegación
- **Diff viewer** — `git diff` mostrado lado a lado o unificado
- **Revert** — Restaurar un request a un commit específico
- **Branch selector** — Cambiar de rama desde la UI
- **Auto-commit toggle** — Opción para desactivar commits automáticos

---

## 5. Historial

### 5.1 Almacenamiento
- Cada petición enviada se guarda en `~/.config/hilo/history/<timestamp>.json`
- Incluye request + response completos
- Metadata: duración, timestamp, código de estado

### 5.2 Navegación
- Lista cronológica (más reciente primero)
- Filtro por método (GET, POST, etc.)
- Filtro por código de estado (2xx, 4xx, 5xx)
- Búsqueda por URL o nombre
- Acciones: **Reusar** (cargar en editor), **Eliminar**, **Guardar en colección**

### 5.3 Límites
- Configurable: máximo de entradas en historial (default: 500)
- Auto-limpieza de los más antiguos al exceder el límite

---

## 6. Gestión de Respuestas

### 6.1 Visualización
- Status code coloreado (2xx verde, 3xx cyan, 4xx amarillo, 5xx rojo)
- Tiempo de respuesta en ms
- Tamaño del body formateado (B, KB, MB)
- Headers de respuesta colapsables

### 6.2 Acciones Post-Response
- **Copiar** — Copiar body/status/headers al portapapeles
- **Guardar** — Guardar respuesta como archivo (JSON, TXT, etc.)
- **Compartir** — Exportar response como snippet
- **Test** — (Futuro) Validar respuesta contra reglas definidas por el usuario

### 6.3 Búsqueda en Respuesta
- `/` para buscar texto dentro del body
- Resultados resaltados
- `n` / `N` para siguiente/anterior match

---

## 7. Importación / Exportación

### 7.1 Importar
- **OpenAPI (Swagger)** — Escanear spec y crear colección con todos los endpoints (Futuro v2)
- **cURL** — Pegar comando cURL y parsear a request
- **Postman Collection** — Importar `collection.json` de Postman
- **HTTPie Session** — Importar desde HTTPie

### 7.2 Exportar
- **cURL** — Copiar como comando cURL
- **OpenAPI** — Exportar colección como spec OpenAPI (Futuro)
- **Raw Request** — Exportar como archivo `.hilo.json`
- **Response** — Guardar respuesta como archivo

---

## 8. Flujos de Trabajo

### 8.1 Flujo Básico
1. Abrir Hilo → Tab Request activo
2. Seleccionar método (GET)
3. Escribir URL: `{{BASE_URL}}/users`
4. Ir a Params → `page=1`, `limit=20`
5. Ir a Headers → `Accept: application/json`
6. Ir a Auth → Bearer → `{{TOKEN}}`
7. `Ctrl+S` → Enviar
8. Ver respuesta en panel inferior
9. Si es útil → `Ctrl+D` para duplicar, modificar, re-enviar
10. Guardar en colección con nombre descriptivo

### 8.2 Flujo con Colecciones
1. Tab Collections → `n` para nueva colección
2. Nombre: "Mi API"
3. Seleccionar entorno: "development"
4. Crear requests dentro de la colección
5. Cada save → commit automático en git
6. Ver historial de cambios → `l` para log
7. Revertir si es necesario

### 8.3 Flujo con Entornos
1. Tab Environments → Crear "development", "staging", "production"
2. En dev: `BASE_URL=https://dev.api.example.com`, `TOKEN=dev-token`
3. En prod: `BASE_URL=https://api.example.com`, `TOKEN=prod-token`
4. En el editor, usar `{{BASE_URL}}` y `{{TOKEN}}`
5. Cambiar de entorno con un atajo → todas las variables se re-resuelven

---

## 9. Glosario de Atajos (Completo)

| Atajo | Contexto | Acción |
|-------|----------|--------|
| `Ctrl+S` | Request | Enviar petición |
| `Ctrl+N` | Request | Nueva petición en blanco |
| `Ctrl+D` | Request | Duplicar petición actual (en la colección activa) |
| `Ctrl+E` | Request | Ciclar sub-tabs (Params/Headers/Auth/Body) |
| `Ctrl+B` | Request | Ciclar tipo de Body (None/JSON/Form/Raw/Binary) |
| `Ctrl+Y` | Request | Copiar petición como cURL al portapapeles |
| `Ctrl+K` | Response | Borrar respuesta |
| `Tab` / `Shift+Tab` | Request | Mover foco (URL → Sub-tabs → Editor → Acciones → Respuesta) |
| `↑`/`↓` | Listas / Editor | Navegar / cambiar método / scroll de respuesta |
| `←`/`→` | Request | Acciones (Send/Save/cURL), columnas del editor, modo de respuesta |
| `Enter` | Listas / Acciones | Seleccionar / abrir / ejecutar acción / añadir fila |
| `Esc` | Any | Atrás / cerrar panel modal |
| `/` | Response | Buscar en respuesta |
| `n` / `N` | Response | Siguiente/anterior match de búsqueda |
| `a` / `Space` | Environments | Activar/desactivar entorno seleccionado |
| `?` | Any | Ayuda contextual |
| `q` | Any (sin editar) | Salir |
| `1`-`6` | Any (sin editar) | Ir directamente al tab N |
| `h` / `l` | Any (sin editar) | Tab anterior / siguiente |

---

## 10. Futuras Extensiones (v2+)

- **WebSocket** — Cliente WebSocket básico
- **GraphQL** — Editor GraphQL con schema introspection
- **gRPC** — Cliente gRPC básico
- **Tests Automáticos** — Definir asserts sobre responses
- **Benchmarking** — Enviar N requests y medir performance
- **Plugins** — Sistema de plugins para transformaciones pre/post request
- **Equipos** — Compartir colecciones via git remoto (push/pull)
- **CLI Mode** — Modo headless para scripting: `hilo send --collection "prod" --request "health"`

---

## 11. Consideraciones Técnicas

### 11.1 Performance
- Respuestas grandes (>1MB) truncadas por defecto en vista, con opción "ver completo"
- Historial con paginación virtual
- Colecciones cargadas lazy (solo metadata hasta que se abre)

### 11.2 Seguridad
- Tokens y passwords **nunca** se muestran completos en UI — mostrar solo últimos 4 chars
- Opción para almacenar valores sensibles en keyring del SO (futuro)
- Git commits: archivo `.gitignore` interno evita hacer commit de archivos con secrets

### 11.3 Resilience
- Timeouts default: 30s por petición
- Errores de red mostrados con claridad (DNS, conexión, TLS, timeout)
- Todos los errores de I/O (disco, git) son manejados silenciosamente (no crashean la TUI)
- Auto-guardado periódico del request en edición

### 11.4 Responsive Layout
- Terminal ≥ 120 columnas: layout completo con editor y response lado a lado
- Terminal ≥ 80 columnas: layout apilado (editor arriba, response abajo)
- Terminal < 80 columnas: modo compacto con paneles intercambiables
