# Radarr Authentication Investigation

**Date**: January 2025  
**Purpose**: Document findings about Radarr's API key authentication system

---

## 🔍 Investigation Summary

**Question**: Can Radarr web UI be auto-logged in using API key in URL?

**Answer**: **NO** - The API key only works for API endpoints, not the web UI.

---

## 📖 Official Radarr Code Analysis

### 1. API Key Authentication Handler

**File**: `src/Radarr.Http/Authentication/ApiKeyAuthenticationHandler.cs`

```csharp
private string ParseApiKey()
{
    // 1. Try query parameter "apikey"
    if (Request.Query.TryGetValue(Options.QueryName, out var value))
    {
        return value.FirstOrDefault();
    }

    // 2. Try header "X-Api-Key"
    if (Request.Headers.TryGetValue(Options.HeaderName, out var headerValue))
    {
        return headerValue.FirstOrDefault();
    }

    // 3. Try Authorization Bearer token
    return Request.Headers["Authorization"].FirstOrDefault()?.Replace("Bearer ", "");
}
```

**Key Points**:
- Radarr accepts API key in **3 formats**:
  1. Query parameter: `?apikey=XXX`
  2. Header: `X-Api-Key: XXX`
  3. Bearer token: `Authorization: Bearer XXX`

### 2. Authentication Configuration

**File**: `src/Radarr.Http/Authentication/AuthenticationBuilderExtensions.cs`

```csharp
.AddApiKey("API", options =>
{
    options.HeaderName = "X-Api-Key";
    options.QueryName = "apikey";  // ✅ Query parameter IS supported
})
```

### 3. Frontend Implementation

**File**: `frontend/src/Utilities/createAjaxRequest.js`

```javascript
function addApiKey(ajaxOptions) {
  ajaxOptions.headers = ajaxOptions.headers || {};
  ajaxOptions.headers['X-Api-Key'] = window.Radarr.apiKey;
}
```

**Observation**: Radarr's own frontend uses **header**, not query parameter.

---

## 🎯 Critical Finding: API Key Scope

### Where API Key Works

✅ **API Endpoints** (`/api/*`):
- `/api/v3/system/status?apikey=XXX` ✅ Works
- `/api/v3/movie?apikey=XXX` ✅ Works
- Any endpoint under `/api/*` ✅ Works

### Where API Key DOES NOT Work

❌ **Web UI Pages**:
- `/?apikey=XXX` ❌ Redirects to login
- `/login?apikey=XXX` ❌ Ignored
- `/movies?apikey=XXX` ❌ Requires authentication

**File**: `src/Radarr.Http/Extensions/RequestExtensions.cs`

```csharp
public static bool IsApiRequest(this HttpRequest request)
{
    return request.Path.StartsWithSegments("/api", StringComparison.InvariantCultureIgnoreCase);
}
```

**Proof**: API key authentication only applies to requests matching `/api/*` path.

---

## 🔒 Radarr Authentication System

### Authentication Types

**File**: `src/NzbDrone.Core/Configuration/ConfigFileProvider.cs`

```xml
<AuthenticationMethod>Forms</AuthenticationMethod>
<AuthenticationRequired>DisabledForLocalAddresses</AuthenticationRequired>
```

### Login Flow

**File**: `src/Radarr.Http/Authentication/AuthenticationController.cs`

```csharp
[HttpPost("login")]
public async Task<IActionResult> Login([FromForm] LoginResource resource, [FromQuery] string returnUrl = null)
{
    var user = _authService.Login(HttpContext.Request, resource.Username, resource.Password);
    
    if (user == null)
    {
        return Redirect($"~/login?returnUrl={returnUrl}&loginFailed=true");
    }
    
    // Create cookie session
    await HttpContext.SignInAsync(AuthenticationType.Forms.ToString(), ...);
    
    return Redirect(_configFileProvider.UrlBase + returnUrl);
}
```

**Key Points**:
- Web UI uses **cookie-based authentication** (Forms)
- Login requires **username + password** (POST form)
- Session stored in **cookie**, not API key
- API key is **separate authentication scheme** for API endpoints only

---

## 🧪 Tested URLs

| URL | Result | Explanation |
|-----|--------|-------------|
| `http://radarr.docker.internal/` | ⚠️ Redirects to login | No authentication |
| `http://radarr.docker.internal/?apikey=XXX` | ⚠️ Redirects to login | API key ignored on UI routes |
| `http://radarr.docker.internal/login?returnUrl=%2F%3Fapikey%3DXXX` | ⚠️ Shows login form | API key in returnUrl ignored |
| `http://radarr.docker.internal/api/v3/system/status?apikey=XXX` | ✅ Works | API endpoint accepts query parameter |

---

## 💡 Possible Workarounds

### Option 1: Accept Manual Login (RECOMMENDED)
- Open Radarr UI directly: `http://radarr.docker.internal/`
- User logs in manually via Radarr's login form
- Session persists via cookie
- **MediaCheky action**: Simply open URL without API key

### Option 2: Programmatic Login + Cookie
- POST credentials to `/login` endpoint
- Extract session cookie from response
- Store cookie in browser
- Redirect to Radarr UI
- **Complexity**: Requires managing Radarr credentials in MediaCheky

### Option 3: iframe with API Proxy
- Embed Radarr in iframe
- Proxy all requests through MediaCheky
- Inject `X-Api-Key` header in proxy
- **Complexity**: High, requires full proxy layer

### Option 4: Radarr Configuration Change
- Disable authentication: `<AuthenticationMethod>None</AuthenticationMethod>`
- **Security risk**: Only for trusted networks
- **Not recommended**: Reduces security

---

## 📝 Implementation Decision

**Chosen**: **Option 1 - Accept Manual Login**

**Reasoning**:
1. **Security**: Don't store Radarr credentials in MediaCheky
2. **Simplicity**: Single `window.open()` call
3. **Official behavior**: Radarr's design separates API key (API) from cookie auth (UI)
4. **User experience**: Login once in Radarr, session persists

**Code Change**:
```javascript
async openRadarrWithAuth() {
    const protocol = this.proxyConfig.sslEnabled ? 'https://' : 'http://';
    const url = `${protocol}${this.serviceEndpoint}`;
    window.open(url, '_blank', 'noopener,noreferrer');
}
```

---

## 🔗 References

- **Radarr Repository**: https://github.com/Radarr/Radarr
- **Authentication Handler**: `src/Radarr.Http/Authentication/ApiKeyAuthenticationHandler.cs`
- **Login Controller**: `src/Radarr.Http/Authentication/AuthenticationController.cs`
- **API Detection**: `src/Radarr.Http/Extensions/RequestExtensions.cs`

---

## 📌 Key Takeaways

1. ✅ Radarr DOES support `?apikey=XXX` in URLs
2. ⚠️ But ONLY for `/api/*` endpoints, NOT web UI
3. 🔒 Web UI requires cookie-based authentication (username/password login)
4. 🎯 MediaCheky opens Radarr UI directly, user logs in manually
5. 📖 This behavior is by design in Radarr's architecture

---

**Last Updated**: January 2025  
**Status**: Investigation Complete ✅

