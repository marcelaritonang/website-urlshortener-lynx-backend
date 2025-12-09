# Lynx URL Shortener - API Documentation

**Base URL:** `http://localhost:8080`  
**Version:** 1.0.0 (Production-Ready)

---

## 🔐 Authentication

### Register

```http
POST /v1/auth/register

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

✅ Password hashed with Argon2id  
✅ Rate limit: 5 attempts/15min

### Login

```http
POST /v1/auth/login

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "..."
}
```

✅ Token expires: 24h  
✅ Rate limit: 5 attempts/15min

### Forgot Password

```http
POST /v1/auth/forgot-password

{
  "email": "user@example.com"
}
```

✅ Email sent via SMTP Gmail  
✅ Token valid: 1 hour  
✅ Rate limit: 1 email/5min

### Reset Password

```http
POST /v1/auth/reset-password

{
  "token": "abc123-from-email",
  "new_password": "NewSecurePass123!"
}
```

✅ Token single-use  
✅ All sessions invalidated

### Logout

```http
POST /v1/api/user/logout
Authorization: Bearer {token}
```

✅ All user sessions invalidated

---

## 🔗 URL Shortener

### Create Short URL (Authenticated)

```http
POST /v1/api/urls
Authorization: Bearer {token}

{
  "long_url": "https://www.youtube.com",
  "custom_short_code": "youtube"  // optional
}
```

**Response:**

```json
{
  "short_code": "youtube",
  "short_url": "http://localhost:8080/urls/youtube",
  "clicks": 0,
  "is_anonymous": false
}
```

✅ Permanent (no expiry)  
✅ Cached in Redis (24h)  
✅ Auto-generate 6-char code if not specified

### Create Anonymous URL (Public)

```http
POST /api/urls

{
  "long_url": "https://www.example.com",
  "custom_short_code": "mylink",  // optional
  "expiry_hours": 168  // default: 7 days
}
```

✅ Auto-expires after 7 days  
✅ No authentication required

### Get User URLs

```http
GET /v1/api/urls?page=1&per_page=10
Authorization: Bearer {token}
```

**Response:**

```json
{
  "urls": [
    {
      "short_code": "youtube",
      "long_url": "https://www.youtube.com",
      "clicks": 15, // Real-time (Redis + DB)
      "qr_codes": {
        "png": "http://localhost:8080/qr/youtube",
        "base64": "http://localhost:8080/qr/youtube/base64"
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 25
  }
}
```

✅ Real-time click counts (Redis + PostgreSQL)  
✅ Pagination support

### Redirect to Long URL

```http
GET /urls/{shortCode}
```

**Response:** `301 Moved Permanently`

```http
Location: https://www.youtube.com
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
```

✅ Cache HIT: 1-5ms  
✅ Cache MISS: 50ms  
✅ Click counter incremented async  
✅ Batch sync to DB every 10 clicks

### Delete URL

```http
DELETE /v1/api/urls/{id}
Authorization: Bearer {token}
```

✅ Hard delete (permanent)  
✅ Cache cleared

---

## 🎨 QR Code

### PNG Image

```http
GET /qr/{shortCode}
```

Returns: PNG image (256x256px)

### Base64 String

```http
GET /qr/{shortCode}/base64
```

**Response:**

```json
{
  "qr_code_base64": "data:image/png;base64,iVBORw0KGgo..."
}
```

✅ Cached in Redis (24h)

---

## 👤 User Profile

### Get Profile

```http
GET /v1/api/user/me
Authorization: Bearer {token}
```

**Response:**

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2025-12-08T10:00:00Z"
}
```

✅ Cached in Redis (1h)

---

## 🛡️ Rate Limiting

| Endpoint            | Limit                | Block Duration              |
| ------------------- | -------------------- | --------------------------- |
| **Global**          | 100 req/min          | 30 min (after 3 violations) |
| **Auth**            | 5 attempts/15min     | 30 min                      |
| **Forgot Password** | 1 req/5min per email | -                           |

**Headers:**

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1765210503
```

---

## ⚠️ Error Codes

| Code  | Error                 | Description           |
| ----- | --------------------- | --------------------- |
| `400` | Bad Request           | Invalid input         |
| `401` | Unauthorized          | Invalid/missing token |
| `404` | Not Found             | URL not found         |
| `409` | Conflict              | Email exists          |
| `429` | Too Many Requests     | Rate limit exceeded   |
| `500` | Internal Server Error | Server error          |

**Error Response:**

```json
{
  "success": false,
  "message": "Error description"
}
```

---

## 💻 Frontend Integration (TypeScript)

```typescript
const API_BASE = "http://localhost:8080";

// Login
const login = async (email: string, password: string) => {
  const res = await fetch(`${API_BASE}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json();
  localStorage.setItem("token", data.data.token);
  return data;
};

// Create Short URL
const createURL = async (longUrl: string) => {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/v1/api/urls`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ long_url: longUrl }),
  });
  return res.json();
};

// Get User URLs
const getURLs = async (page = 1) => {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/v1/api/urls?page=${page}&per_page=10`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.json();
};
```

---

## 📊 Performance

- **Redis Cache HIT:** 1-5ms
- **PostgreSQL Query:** 30-50ms
- **Click Counter:** Batch sync every 10 clicks (90% DB write reduction)
- **Cache Strategy:** URL (24h), User (1h), Session (24h)

---

## 🔧 Environment Setup

```env
PORT=8080
DB_HOST=127.0.0.1
DB_NAME=lynx_db
REDIS_HOST=127.0.0.1
JWT_SECRET=your-64-char-secret
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FRONTEND_URL=http://localhost:3000
```

---

## 🎯 Testing (cURL)

```bash
# Register
curl -X POST http://localhost:8080/v1/auth/register \
  -d '{"email":"test@example.com","password":"Test123!"}'

# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -d '{"email":"test@example.com","password":"Test123!"}'

# Create URL
curl -X POST http://localhost:8080/v1/api/urls \
  -H "Authorization: Bearer TOKEN" \
  -d '{"long_url":"https://youtube.com"}'

# Redirect
curl -I http://localhost:8080/urls/shortcode
```

---

**Status:** Production-Ready ✅  
**Security:** Argon2 + JWT + Rate Limiting  
**Performance:** Hybrid Cache (Redis + PostgreSQL)  
**Last Updated:** December 8, 2025
