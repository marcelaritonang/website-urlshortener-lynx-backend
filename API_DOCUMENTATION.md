# URL Shortener API Documentation

Base URL: `http://localhost:8080`

## 📧 Authentication & Password Reset APIs

### 1. Register User

**POST** `/v1/auth/register`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Success Response (201):**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Error Response (409 - Email exists):**

```json
{
  "success": false,
  "message": "user already exists"
}
```

---

### 2. Login

**POST** `/v1/auth/login`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "Password123!"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Error Response (401):**

```json
{
  "success": false,
  "message": "invalid credentials"
}
```

---

### 3. Forgot Password (Request Reset)

**POST** `/v1/auth/forgot-password`

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Password reset email has been sent successfully",
  "data": null
}
```

**Notes:**

- ✅ Email akan dikirim ke Gmail user
- ✅ Token berlaku 1 jam
- ✅ Security: Selalu return success meskipun email tidak ada

---

### 4. Reset Password (Confirm with Token)

**POST** `/v1/auth/reset-password`

**Request Body:**

```json
{
  "token": "d31150af-bd2e-4fbc-86b1-61673bb130e9",
  "new_password": "NewPassword123!"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Password has been reset successfully",
  "data": null
}
```

**Error Response (400 - Invalid/Expired Token):**

```json
{
  "success": false,
  "message": "invalid or expired reset token"
}
```

**Password Requirements:**

- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

---

### 5. Get User Details (Protected)

**GET** `/v1/api/user/me`

**Headers:**

```
Authorization: Bearer {token}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "User details retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 6. Logout (Protected)

**POST** `/v1/api/user/logout`

**Headers:**

```
Authorization: Bearer {token}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Logged out successfully",
  "data": null
}
```

---

## 🔗 URL Shortener APIs

### 7. Create Short URL (Protected)

**POST** `/v1/api/urls`

**Headers:**

```
Authorization: Bearer {token}
```

**Request Body:**

```json
{
  "long_url": "https://www.example.com/very/long/url/path",
  "custom_short_code": "mylink" // optional
}
```

**Success Response (201):**

```json
{
  "success": true,
  "message": "Short URL created successfully",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "short_url": "http://localhost:8080/urls/mylink",
    "short_code": "mylink",
    "long_url": "https://www.example.com/very/long/url/path",
    "clicks": 0,
    "is_anonymous": false,
    "expires_at": null,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 8. Create Anonymous URL (Public)

**POST** `/api/urls`

**No authentication required**

**Request Body:**

```json
{
  "long_url": "https://www.example.com/very/long/url/path",
  "custom_short_code": "temp123", // optional
  "expiry_hours": 24 // optional, default: 168 (7 days)
}
```

**Success Response (201):**

```json
{
  "success": true,
  "message": "Anonymous short URL created successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "short_url": "http://localhost:8080/urls/temp123",
    "short_code": "temp123",
    "long_url": "https://www.example.com/very/long/url/path",
    "clicks": 0,
    "is_anonymous": true,
    "expires_at": "2024-01-16T10:30:00Z",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 9. Get User URLs (Protected)

**GET** `/v1/api/urls?page=1&per_page=10`

**Headers:**

```
Authorization: Bearer {token}
```

**Query Parameters:**

- `page` (optional, default: 1)
- `per_page` (optional, default: 10)

**Success Response (200):**

```json
{
  "success": true,
  "message": "URLs retrieved successfully",
  "data": {
    "urls": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "short_url": "http://localhost:8080/urls/abc123",
        "short_code": "abc123",
        "long_url": "https://www.example.com/page1",
        "clicks": 42,
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "per_page": 10
  }
}
```

---

### 10. Redirect to Long URL (Public)

**GET** `/urls/:shortCode`

**Example:** `http://localhost:8080/urls/abc123`

**Response:**

- **Success:** HTTP 302 Redirect to long URL
- **Error (404):** Short code not found or expired

---

### 11. Delete URL (Protected)

**DELETE** `/v1/api/urls/:id`

**Headers:**

```
Authorization: Bearer {token}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "URL deleted successfully",
  "data": null
}
```

---

## 🎨 QR Code APIs

### 12. Get QR Code Image (Public)

**GET** `/qr/:shortCode`

**Example:** `http://localhost:8080/qr/abc123`

**Response:**

- Content-Type: `image/png`
- Binary image data

---

### 13. Get QR Code Base64 (Public)

**GET** `/qr/:shortCode/base64`

**Success Response (200):**

```json
{
  "success": true,
  "data": {
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

---

## 🔥 Frontend Integration Examples

### React/Next.js Example

```typescript
// filepath: frontend/src/lib/api.ts
const API_BASE_URL = "http://localhost:8080";

// Forgot Password Request
export async function forgotPassword(email: string) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/forgot-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  return response.json();
}

// Reset Password
export async function resetPassword(token: string, newPassword: string) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/reset-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      token,
      new_password: newPassword,
    }),
  });
  return response.json();
}

// Login
export async function login(email: string, password: string) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return response.json();
}

// Create Short URL (Authenticated)
export async function createShortURL(
  longUrl: string,
  token: string,
  customCode?: string
) {
  const response = await fetch(`${API_BASE_URL}/v1/api/urls`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      long_url: longUrl,
      custom_short_code: customCode,
    }),
  });
  return response.json();
}

// Create Anonymous URL (No auth)
export async function createAnonymousURL(
  longUrl: string,
  expiryHours?: number
) {
  const response = await fetch(`${API_BASE_URL}/api/urls`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      long_url: longUrl,
      expiry_hours: expiryHours || 168,
    }),
  });
  return response.json();
}
```

---

## 📝 Response Format

All API responses follow this format:

**Success Response:**

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    /* response data */
  }
}
```

**Error Response:**

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

---

## 🔒 Authentication

Protected endpoints require JWT token in header:

```
Authorization: Bearer {your_jwt_token}
```

Token expires after 24 hours. Use refresh token to get new access token.

---

## 🌐 CORS

Frontend allowed origins: `*` (configure in production)

---

## 📧 Email Flow (Forgot Password)

1. **User requests password reset** → `POST /v1/auth/forgot-password`
2. **System sends email** to user's Gmail with reset link
3. **Email contains:** `http://localhost:3000/reset-password?token=...`
4. **User clicks link** → Frontend displays reset password form
5. **User submits new password** → `POST /v1/auth/reset-password`
6. **Password updated** → User can login with new password

---

## 🎯 Testing with Postman/cURL

### Forgot Password

```bash
curl -X POST http://localhost:8080/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

### Reset Password

```bash
curl -X POST http://localhost:8080/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token":"d31150af-bd2e-4fbc-86b1-61673bb130e9",
    "new_password":"NewPassword123!"
  }'
```

---

## ⚠️ Important Notes

1. **Password Requirements:** Min 8 chars, uppercase, lowercase, number, special char
2. **Reset Token:** Valid for 1 hour only
3. **Token Usage:** Single-use token (deleted after successful reset)
4. **Anonymous URLs:** Default expiry 7 days (168 hours)
5. **Authenticated URLs:** No expiry (permanent until deleted)
6. **Email Security:** Always returns success even if email doesn't exist

---

## 🚀 Quick Start for Frontend Developers

```typescript
// 1. Request password reset
const result = await forgotPassword("user@example.com");
// Check email for reset link

// 2. User opens link: /reset-password?token=xxx
// Extract token from URL query params

// 3. Submit new password
const resetResult = await resetPassword(token, "NewPassword123!");
if (resetResult.success) {
  // Redirect to login page
  router.push("/login");
}
```

---

**Backend Ready ✅** | **API Version:** 1.0.0 | **Last Updated:** 2024-01-15
