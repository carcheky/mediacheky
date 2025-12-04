# Testing the Radarr Authentication API Integration

This guide walks through testing the new Radarr authentication API integration that was just implemented.

## Prerequisites

1. MediaCheky running with `make dev`
2. Radarr service running and accessible
3. Radarr API key available
4. Browser with developer console access (Chrome, Firefox, etc.)

## Setup

### 1. Configure Environment Variable

Ensure `RADARR_API_KEY` is set in your environment:

```bash
# In docker-compose.yml or .env
RADARR_API_KEY=your-radarr-api-key-here
```

Get your Radarr API key from:
- Radarr UI → Settings → General → API Key

### 2. Verify Services are Running

```bash
# Check that both services are running
docker compose ps

# Should show:
# mediacheky    Up
# radarr        Up
```

## Testing Scenarios

### Test 1: Load Authentication Settings (Automatic)

**Expected Behavior:** When navigating to Service Config for Radarr, authentication settings load automatically.

**Steps:**
1. Navigate to `http://localhost:7369/settings`
2. Click on Radarr "Configure" button
3. Wait for page to load (3-5 seconds)
4. Check browser console (F12)

**Verification:**
- [ ] Username field is populated with current Radarr username
- [ ] Console shows no JavaScript errors
- [ ] Network tab shows successful request to `/api/services/radarr/radarr/auth`
- [ ] Response contains `"success": true`
- [ ] Toast notification does NOT appear (because it loaded in background)

**Expected Response:**
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

### Test 2: Update Authentication Settings

**Expected Behavior:** User can update Radarr credentials and changes are saved.

**Steps:**
1. On the Service Config page, scroll to "Security" section
2. Find the "auth API" section (different from app config section)
3. Enter new username: `testuser`
4. Enter new password: `testpass123`
5. Click "Update Credentials" button
6. Wait for operation to complete

**Verification:**
- [ ] Button shows loading spinner during update
- [ ] Success toast notification appears: "Authentication updated successfully!"
- [ ] Password field is cleared after successful update
- [ ] Network tab shows PUT request to `/api/services/radarr/radarr/auth`
- [ ] Request body contains new username and password
- [ ] Response contains `"success": true`

**Expected Request Body:**
```json
{
  "id": 1,
  "method": "Forms",
  "authenticationRequired": "DisabledForLocalAddresses",
  "authenticationMethod": "Forms",
  "username": "testuser",
  "password": "testpass123"
}
```

### Test 3: Refresh Page to Verify Save

**Expected Behavior:** After saving, credentials should persist in Radarr.

**Steps:**
1. After successful update, reload the page (F5)
2. Wait for Service Config to load
3. Check username field

**Verification:**
- [ ] Username field now shows the new value (`testuser`)
- [ ] Page loads without errors
- [ ] Auth settings loaded successfully from Radarr

### Test 4: Error Handling - Empty Fields

**Expected Behavior:** System should validate and reject empty credentials.

**Steps:**
1. Clear the username field
2. Keep password filled
3. Click "Update Credentials"

**Verification:**
- [ ] Error toast appears: "Username is required"
- [ ] No request sent to backend (caught by frontend validation)
- [ ] Button returns to normal state

### Test 5: Error Handling - Wrong API Key

**Expected Behavior:** System handles API authentication errors gracefully.

**Steps:**
1. Temporarily change `RADARR_API_KEY` to invalid value
2. Restart MediaCheky: `make dev`
3. Try to load Service Config for Radarr
4. Or try to update credentials

**Verification:**
- [ ] Error toast appears with user-friendly message
- [ ] Error message indicates authentication issue (e.g., "Invalid API key")
- [ ] Page doesn't crash
- [ ] Error appears in backend logs

### Test 6: Error Handling - Radarr Not Running

**Expected Behavior:** System handles Radarr connection failures gracefully.

**Steps:**
1. Stop Radarr: `docker compose stop radarr`
2. Try to load Service Config for Radarr (or reload existing page)
3. Or try to update credentials

**Verification:**
- [ ] Error toast appears with connection error message
- [ ] Message indicates "Radarr" is unreachable or connection refused
- [ ] Page doesn't crash
- [ ] Error appears in backend logs

### Test 7: UI Elements Verification

**Expected Behavior:** All UI elements display correctly and are styled consistently.

**Steps:**
1. Navigate to Service Config for Radarr
2. Scroll to Security section
3. Look for authentication fields

**Verification:**
- [ ] Username field labeled "Username" with "auth API" badge
- [ ] Password field labeled "Password" with "auth API" badge
- [ ] Both fields have helper text about API authentication
- [ ] "Update Credentials" button is amber colored (different from other buttons)
- [ ] Button has hover effect
- [ ] Loading spinner appears when clicked
- [ ] Form is responsive on mobile screen sizes

## API Endpoint Testing

### Using curl to Test GET Endpoint

```bash
# Get authentication settings
curl -X GET http://localhost/api/services/radarr/radarr/auth \
  -H "Content-Type: application/json"
```

### Using curl to Test PUT Endpoint

```bash
# Update authentication settings
curl -X PUT http://localhost/api/services/radarr/radarr/auth \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "method": "Forms",
    "authenticationRequired": "DisabledForLocalAddresses",
    "authenticationMethod": "Forms",
    "username": "newuser",
    "password": "newpass"
  }'
```

## Running Automated Tests

### Unit Tests

```bash
# Run all Radarr auth tests
go test ./internal/handler -run TestRadarrAuth -v

# Run specific test
go test ./internal/handler -run TestRadarrAuthConfig -v

# Run with coverage
go test ./internal/handler -run TestRadarrAuth -cover
```

### Test Output

Expected output:
```
=== RUN   TestRadarrAuthConfig
=== RUN   TestRadarrAuthConfig/successful_auth_config_retrieval
--- PASS: TestRadarrAuthConfig (0.00s)
    --- PASS: TestRadarrAuthConfig/successful_auth_config_retrieval (0.00s)

=== RUN   TestUpdateRadarrAuthConfig
=== RUN   TestUpdateRadarrAuthConfig/successful_auth_config_update
--- PASS: TestUpdateRadarrAuthConfig (0.00s)
    --- PASS: TestUpdateRadarrAuthConfig/successful_auth_config_update (0.00s)

PASS
ok      github.com/carcheky/mediacheky/internal/handler    0.006s
```

## Browser Console Checks

### Expected Console Logs

When loading Service Config for Radarr, you should see:
```javascript
// No errors - clean load
// Network request successful
// Methods loaded: loadRadarrAuthSettings(), saveRadarrAuthSettings()
```

### Debugging with Console

```javascript
// In browser console, verify methods exist
typeof serviceConfig.loadRadarrAuthSettings  // Should be: "function"
typeof serviceConfig.saveRadarrAuthSettings  // Should be: "function"

// Manually trigger load
serviceConfig.loadRadarrAuthSettings()

// Check form values
document.querySelector('#radarr-username').value
document.querySelector('#radarr-password').value
```

## Integration Testing Checklist

- [ ] **Load Settings**
  - Settings load automatically on page init
  - Username field populated correctly
  - No JavaScript errors in console

- [ ] **Update Credentials**
  - Button click triggers update
  - Loading spinner appears
  - Success message shown
  - Credentials actually changed in Radarr

- [ ] **Error Handling**
  - Empty fields validated
  - Connection errors handled
  - API key errors handled
  - User-friendly error messages shown

- [ ] **Security**
  - Password field cleared after save
  - Password never shown to user on load
  - API key not exposed in logs or requests
  - HTTPS should be used in production

- [ ] **UI/UX**
  - All elements properly styled
  - Loading states work correctly
  - Toast notifications appear and auto-dismiss
  - Form is responsive on different screen sizes

## Troubleshooting

### If settings don't load:
1. Check browser console for JavaScript errors
2. Verify RADARR_API_KEY is set correctly
3. Check that Radarr is running: `docker compose ps radarr`
4. Check backend logs: `docker compose logs mediacheky`

### If update fails:
1. Verify new credentials are correct in Radarr UI
2. Check error message in toast notification
3. Look at backend logs for detailed error
4. Try with simpler credentials first (no special chars)

### If password field not clearing:
1. Check error message - save may have failed
2. Check browser console for JavaScript errors
3. Reload page to verify change was saved

## Success Criteria

The implementation is considered successful when:

1. ✅ Settings load automatically without user action
2. ✅ Users can update credentials through the UI
3. ✅ Changes are reflected in Radarr immediately
4. ✅ All error scenarios are handled gracefully
5. ✅ UI is intuitive and responsive
6. ✅ All tests pass
7. ✅ No JavaScript errors in console
8. ✅ No breaking changes to existing functionality

## Documentation

- Complete API documentation: `docs/RADARR_AUTH_INTEGRATION.md`
- Implementation details: `docs/RADARR_AUTH_IMPLEMENTATION.md`
- Tests: `internal/handler/radarr_auth_test.go`

