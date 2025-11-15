# Radarr Authentication Integration Guide

This document describes how to use the new Radarr authentication API integration in MediaCheky.

## Overview

MediaCheky now provides a unified interface for managing Radarr authentication credentials. This allows users to:
- Load current Radarr authentication settings from the Radarr API
- Update authentication username and password
- Validate credentials before saving
- Handle errors gracefully with user-friendly messages

## API Endpoints

### GET /api/services/:name/radarr/auth

Retrieves current Radarr authentication configuration.

**Parameters:**
- `:name` - Service name (typically "radarr")

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "method": "Forms",
    "authenticationRequired": "DisabledForLocalAddresses",
    "authenticationMethod": "Forms",
    "username": "admin",
    "passwordField": "password"
  },
  "error": ""
}
```

**Error Response:**
```json
{
  "success": false,
  "data": null,
  "error": "Unable to connect to Radarr: connection refused"
}
```

### PUT /api/services/:name/radarr/auth

Updates Radarr authentication configuration.

**Parameters:**
- `:name` - Service name (typically "radarr")

**Request Body:**
```json
{
  "id": 1,
  "method": "Forms",
  "authenticationRequired": "DisabledForLocalAddresses",
  "authenticationMethod": "Forms",
  "username": "newuser",
  "password": "newpassword"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "method": "Forms",
    "authenticationRequired": "DisabledForLocalAddresses",
    "authenticationMethod": "Forms",
    "username": "newuser"
  },
  "error": ""
}
```

## Frontend Integration

### Loading Authentication Settings

```javascript
async loadRadarrAuthSettings() {
    // Load current settings from the backend
    const response = await fetch(`/api/services/${this.serviceName}/radarr/auth`);
    const result = await response.json();
    
    // Populate form fields
    if (result.data.username) {
        document.querySelector('#radarr-username').value = result.data.username;
    }
}
```

### Saving Authentication Settings

```javascript
async saveRadarrAuthSettings() {
    // Get form values
    const username = document.querySelector('#radarr-username').value;
    const password = document.querySelector('#radarr-password').value;
    
    // Load current config first
    const getResponse = await fetch(`/api/services/${this.serviceName}/radarr/auth`);
    const currentConfig = await getResponse.json();
    
    // Update with new values
    currentConfig.data.username = username;
    currentConfig.data.password = password;
    
    // Send update
    const putResponse = await fetch(`/api/services/${this.serviceName}/radarr/auth`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(currentConfig.data)
    });
}
```

## Backend Handler Implementation

### RadarrAuthConfig Handler (GET)

Retrieves authentication configuration from Radarr:

1. Extract API key from environment variables
2. Make HTTP request to `http://radarr:7878/api/v3/config/auth`
3. Add `X-Api-Key` header for authentication
4. Return configuration or error

### UpdateRadarrAuthConfig Handler (PUT)

Updates authentication configuration in Radarr:

1. Extract API key from environment variables
2. Validate request body
3. Make HTTP PUT request to `http://radarr:7878/api/v3/config/auth`
4. Send updated configuration
5. Return success or error

## Environment Variables

The backend uses the following environment variable to authenticate with Radarr:

```bash
RADARR_API_KEY=your-api-key-here
```

This should be set in your `docker-compose.yml` or `.env` file.

## Error Handling

### Common Errors

#### Connection Failed
**Cause:** Radarr service is not accessible
**Solution:** 
1. Verify Radarr is running: `docker compose ps radarr`
2. Check Radarr URL is correct
3. Verify network connectivity

#### Invalid API Key
**Cause:** RADARR_API_KEY environment variable is incorrect or not set
**Solution:**
1. Get API key from Radarr: Settings > General > API Key
2. Update environment variable
3. Restart MediaCheky service

#### Authentication Failed
**Cause:** New credentials are incorrect or already in use
**Solution:**
1. Verify credentials are correct
2. Check Radarr logs for authentication errors
3. Try resetting credentials in Radarr UI

### Error Response Example

```json
{
  "success": false,
  "data": null,
  "error": "Failed to connect to Radarr: connection timeout after 10 seconds"
}
```

## Security Considerations

1. **Password Handling**
   - Passwords are never displayed to users (only sent to Radarr)
   - When loading config, password field is returned empty for security
   - Always use HTTPS in production

2. **API Key Storage**
   - API key is read from environment variables only
   - Never stored in database or logs
   - Should be protected like a password

3. **Request/Response Security**
   - All requests to Radarr API use X-Api-Key authentication
   - Responses are validated before use
   - Errors are user-friendly but don't leak technical details

## Testing

Run the included test suite:

```bash
go test ./internal/handler -run TestRadarrAuthConfig
go test ./internal/handler -run TestUpdateRadarrAuthConfig
go test ./internal/handler -run TestExtractRadarrAPIKey
```

## UI Workflow

1. **Load Settings**
   - User navigates to Service Config page
   - init() automatically loads current Radarr auth settings
   - Username field is populated (password remains empty)

2. **Update Credentials**
   - User enters new username in "#radarr-username" field
   - User enters new password in "#radarr-password" field
   - User clicks "Update Credentials" button

3. **Save and Validate**
   - Frontend calls saveRadarrAuthSettings()
   - Validation ensures both fields are filled
   - Backend calls Radarr API with new credentials
   - Toast notification shows success or error

4. **Post-Update**
   - Password field is cleared for security
   - Success message tells user they may need to re-login
   - Settings remain loaded for further modifications

## Troubleshooting

### Settings not loading
1. Check browser console for JavaScript errors
2. Verify API endpoint is accessible: `curl http://localhost:7369/api/services/radarr/radarr/auth`
3. Check backend logs: `docker compose logs mediacheky`

### Updates not saving
1. Verify Radarr service is running
2. Check API key is correct in environment
3. Review error message in toast notification
4. Check Radarr logs for validation errors

### Password not clearing after update
1. Check if error occurred (error message shown)
2. If successful, manually clear using form reset button
3. Verify update was actually successful by reloading page

## API Documentation References

- [Radarr API v3 Documentation](https://radarr.servarr.com/docs/api/)
- [Config/Auth Endpoint](https://radarr.servarr.com/docs/api/#/config/get_api_v3_config_auth)

## Future Enhancements

- [ ] Support for Basic Authentication
- [ ] Support for External Authentication
- [ ] Password strength validation
- [ ] Two-factor authentication support
- [ ] Authentication history/audit log
- [ ] Bulk authentication updates for multiple services

