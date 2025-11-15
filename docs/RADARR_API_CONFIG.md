# Radarr API Configuration Guide

## Overview

Esta guía documenta cómo interactuar con la API de Radarr para guardar y actualizar configuraciones desde MediaCheky.

## Autenticación

Todos los requests a la API de Radarr requieren:
- **API Key**: Se envía como parámetro de query o header
- **Header**: `X-Api-Key: {apiKey}`
- **URL Base**: `http://radarr-host:7878` (puerto por defecto, varía según configuración)

## Endpoints Principales

### 1. Obtener Configuración de Seguridad/Autenticación

**GET** `/api/v3/config/auth`

**Parámetros**: Requiere API Key

**Respuesta**:
```json
{
  "id": 1,
  "method": "forms",
  "authenticationRequired": "enabled|disabledForLocalAdresses|disabled",
  "authenticationMethod": "forms|basic|oauth",
  "username": "admin",
  "password": "hashed_password_not_returned"
}
```

**Nota**: La contraseña nunca se retorna en el GET, solo se puede establecer en POST/PUT.

### 2. Actualizar Configuración de Seguridad

**PUT** `/api/v3/config/auth`

**Headers**:
```
X-Api-Key: {apiKey}
Content-Type: application/json
```

**Body**:
```json
{
  "id": 1,
  "method": "forms",
  "authenticationRequired": "enabled",
  "authenticationMethod": "forms",
  "username": "newusername",
  "password": "newpassword"
}
```

**Respuesta**: Retorna el objeto actualizado

**Importante**: 
- Este endpoint requiere autenticación PREVIA
- Si cambias credenciales, después ya no tendrás acceso con las credenciales antiguas
- Radarr hace hash de la contraseña antes de guardarla

### 3. Obtener Todas las Configuraciones

**GET** `/api/v3/config`

Retorna todas las configuraciones disponibles (naming, media management, notifications, etc.)

### 4. Configuración de Paths

**PUT** `/api/v3/config/mediamanagement`

Actualiza rutas de películas, descargas, etc.

```json
{
  "id": 1,
  "moviePathFormat": "{Movie Title} ({Release Year})",
  "movieFolderFormat": "{Movie Title} ({Release Year})",
  "multiEpisodeStyle": 1,
  "minorMoviePathFormat": "{Movie Title} ({Release Year})",
  "createEmptySeriesFolders": false,
  "deleteEmptyFolders": true,
  "enableMediaInfo": true,
  "enableRecycleBin": true,
  "recycleBinCleanupDays": 7,
  "firstDayOfWeek": 0,
  "weekColumnHeader": "ddd M/d",
  "shortDateFormat": "YYYY-MM-DD",
  "longDateFormat": "dddd, MMMM D, YYYY",
  "timeFormat": "h(:mm)a",
  "showRelativeDates": true,
  "enableColorImpairedMode": false
}
```

## Flujo de Actualización desde MediaCheky

### Paso 1: Validar Conexión

```bash
curl -H "X-Api-Key: YOUR_API_KEY" http://radarr:7878/api/v3/system/status
```

### Paso 2: Obtener Configuración Actual

```bash
curl -H "X-Api-Key: YOUR_API_KEY" http://radarr:7878/api/v3/config/auth
```

### Paso 3: Enviar Actualización

```bash
curl -X PUT \
  -H "X-Api-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "method": "forms",
    "authenticationRequired": "enabled",
    "authenticationMethod": "forms",
    "username": "newuser",
    "password": "newpass"
  }' \
  http://radarr:7878/api/v3/config/auth
```

## Consideraciones de Seguridad

⚠️ **IMPORTANTE**:

1. **Cambio de Credenciales**: Si cambias el usuario/contraseña en Radarr vía API, necesitarás:
   - Las NUEVAS credenciales para futuros requests
   - Actualizar la configuración en MediaCheky con las nuevas credenciales

2. **Permisos de API**: 
   - Solo un usuario con acceso admin puede cambiar configuraciones sensibles
   - La API Key debe tener permisos completos

3. **HTTPS en Producción**:
   - Siempre usa HTTPS cuando accedas a Radarr desde internet
   - Nunca transmitas credenciales en texto plano

4. **Validación**:
   - Siempre valida el formato de entrada antes de enviar
   - Guarda la configuración anterior para poder hacer rollback si falla

## Endpoints Relacionados

### Obtener Status
**GET** `/api/v3/system/status`

Verifica que Radarr está ejecutando y accesible.

### Obtener Versión
**GET** `/api/v3/system/status`

Incluye la versión de Radarr en la respuesta.

### Health Check
**GET** `/api/v3/health`

Retorna información de salud de Radarr.

## Ejemplo Completo: Frontend (Alpine.js)

```javascript
async function saveRadarrAuth() {
    const username = document.querySelector('#radarr-username').value;
    const password = document.querySelector('#radarr-password').value;
    
    try {
        // Primero obtener la configuración actual
        const getResp = await fetch(`/api/radarr/auth`, {
            headers: {
                'X-Api-Key': this.radarrApiKey
            }
        });
        
        if (!getResp.ok) throw new Error('Failed to fetch current config');
        
        const config = await getResp.json();
        
        // Actualizar con nuevos valores
        config.username = username;
        config.password = password;
        
        // Enviar actualización
        const putResp = await fetch(`/api/radarr/auth`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-Api-Key': this.radarrApiKey
            },
            body: JSON.stringify(config)
        });
        
        if (!putResp.ok) throw new Error('Failed to update auth');
        
        this.showToast('Authentication updated successfully', 'success');
    } catch (error) {
        this.showToast(`Error: ${error.message}`, 'error');
    }
}
```

## Ejemplo Backend (Go)

```go
// Obtener configuración de autenticación de Radarr
func (h *RadarrHandler) GetAuthConfig(c *fiber.Ctx) error {
    serviceName := c.Params("name")
    
    // Obtener endpoint de Radarr
    service, err := h.repo.GetService(serviceName)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Llamar a Radarr API
    resp, err := http.Get(fmt.Sprintf("%s/api/v3/config/auth", service.Endpoint))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    defer resp.Body.Close()
    
    var authConfig map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&authConfig); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(authConfig)
}

// Actualizar configuración de autenticación de Radarr
func (h *RadarrHandler) UpdateAuthConfig(c *fiber.Ctx) error {
    serviceName := c.Params("name")
    
    var req map[string]interface{}
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }
    
    // Obtener servicio y su API key
    service, err := h.repo.GetService(serviceName)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Preparar request a Radarr
    bodyBytes, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("PUT", 
        fmt.Sprintf("%s/api/v3/config/auth", service.Endpoint),
        bytes.NewReader(bodyBytes))
    
    httpReq.Header.Set("X-Api-Key", service.ApiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    // Ejecutar request
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return c.Status(resp.StatusCode).JSON(fiber.Map{
            "error": fmt.Sprintf("Radarr API returned %d", resp.StatusCode),
        })
    }
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    return c.JSON(result)
}
```

## Referencias

- [Radarr GitHub](https://github.com/Radarr/Radarr)
- [Radarr API v3 Documentation](https://radarr.video/docs/api/)
- [Servarr Wiki](https://wiki.servarr.com/radarr)
- [Similar Pattern: Sonarr API](https://sonarr.tv/docs/api/)

## Notas de Implementación

- Los endpoints `/api/v3/config/*` requieren autenticación previa
- Radarr guarda contraseñas en texto plano en `config.xml` (cifrado internamente)
- Después de cambiar credenciales, Radarr NO requiere restart
- Todos los cambios se aplican inmediatamente
- Se recomienda hacer GET primero para obtener el `id` antes de hacer PUT
